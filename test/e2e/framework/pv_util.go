/*
Copyright 2015 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package framework

import (
	"fmt"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ec2"
	. "github.com/onsi/ginkgo"
	"google.golang.org/api/googleapi"
	apierrs "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/v1"
	"k8s.io/kubernetes/pkg/client/clientset_generated/clientset"
	awscloud "k8s.io/kubernetes/pkg/cloudprovider/providers/aws"
	gcecloud "k8s.io/kubernetes/pkg/cloudprovider/providers/gce"
	"k8s.io/kubernetes/pkg/volume/util/volumehelper"
)

const (
	PDRetryTimeout    = 5 * time.Minute
	PDRetryPollTime   = 5 * time.Second
	VolumeSelectorKey = "e2e-pv-pool"
)

// Map of all PVs used in the multi pv-pvc tests. The key is the PV's name, which is
// guaranteed to be unique. The value is {} (empty struct) since we're only interested
// in the PV's name and if it is present. We must always Get the pv object before
// referencing any of its values, eg its ClaimRef.
type pvval struct{}
type PVMap map[string]pvval

// Map of all PVCs used in the multi pv-pvc tests. The key is "namespace/pvc.Name". The
// value is {} (empty struct) since we're only interested in the PVC's name and if it is
// present. We must always Get the pvc object before referencing any of its values, eg.
// its VolumeName.
// Note: It's unsafe to add keys to a map in a loop. Their insertion in the map is
//   unpredictable and can result in the same key being iterated over again.
type pvcval struct{}
type PVCMap map[types.NamespacedName]pvcval

// PersistentVolumeConfig is consumed by MakePersistentVolume() to generate a PV object
// for varying storage options (NFS, ceph, glusterFS, etc.).
// (+optional) prebind holds a pre-bound PVC
// Example pvSource:
//	pvSource: api.PersistentVolumeSource{
//		NFS: &api.NFSVolumeSource{
//	 		...
//	 	},
//	 }
type PersistentVolumeConfig struct {
	PVSource      v1.PersistentVolumeSource
	Prebind       *v1.PersistentVolumeClaim
	ReclaimPolicy v1.PersistentVolumeReclaimPolicy
	NamePrefix    string
	Labels        labels.Set
}

// PersistentVolumeClaimConfig is consumed by MakePersistentVolumeClaim() to generate a PVC object.
// AccessModes defaults to all modes (RWO, RWX, ROX) if left empty
// (+optional) Annotations defines the PVC's annotations

type PersistentVolumeClaimConfig struct {
	AccessModes []v1.PersistentVolumeAccessMode
	Annotations map[string]string
	Selector    *metav1.LabelSelector
}

// Clean up a pv and pvc in a single pv/pvc test case.
func PVPVCCleanup(c clientset.Interface, ns string, pv *v1.PersistentVolume, pvc *v1.PersistentVolumeClaim) error {

	var err error
	if pvc != nil {
		if err = DeletePersistentVolumeClaim(c, pvc.Name, ns); err != nil {
			return err
		}
	}
	if pv != nil {
		if err = DeletePersistentVolume(c, pv.Name); err != nil {
			return err
		}
	}
	return nil
}

// Clean up pvs and pvcs in multi-pv-pvc test cases. All entries found in the pv and
// claims maps are deleted.
func PVPVCMapCleanup(c clientset.Interface, ns string, pvols PVMap, claims PVCMap) error {

	var err error

	for pvcKey := range claims {
		if err = DeletePersistentVolumeClaim(c, pvcKey.Name, ns); err != nil {
			return fmt.Errorf("Failed to delete PVC %q: %v", pvcKey.Name, err)
		}
		delete(claims, pvcKey)
	}

	for pvKey := range pvols {
		if err = DeletePersistentVolume(c, pvKey); err != nil {
			return fmt.Errorf("Failed to delete PV %q: %v", pvKey, err)
		}
		delete(pvols, pvKey)
	}
	return nil
}

// Delete the PV.
func DeletePersistentVolume(c clientset.Interface, pvName string) error {
	if c != nil && len(pvName) > 0 {
		Logf("Deleting PersistentVolume %q", pvName)
		err := c.CoreV1().PersistentVolumes().Delete(pvName, nil)
		if err != nil && !apierrs.IsNotFound(err) {
			return fmt.Errorf("PV Delete API error: %v", err)
		}
	}
	return nil
}

// Delete the Claim
func DeletePersistentVolumeClaim(c clientset.Interface, pvcName string, ns string) error {
	if c != nil && len(pvcName) > 0 {
		Logf("Deleting PersistentVolumeClaim %v", pvcName)
		err := c.CoreV1().PersistentVolumeClaims(ns).Delete(pvcName, nil)
		if err != nil && !apierrs.IsNotFound(err) {
			return fmt.Errorf("PVC Delete API error: %v", err)
		}
	}
	return nil
}

// Delete the PVC and wait for the PV to enter its expected phase. Validate that the PV
// has been reclaimed (assumption here about reclaimPolicy). Caller tells this func which
// phase value to expect for the pv bound to the to-be-deleted claim.
func DeletePVCandValidatePV(c clientset.Interface, ns string, pvc *v1.PersistentVolumeClaim, pv *v1.PersistentVolume, expectPVPhase v1.PersistentVolumePhase) error {

	var err error

	pvname := pvc.Spec.VolumeName
	Logf("Deleting PVC %v to trigger reclamation of PV %v", pvc.Name, pvname)
	if err = DeletePersistentVolumeClaim(c, pvc.Name, ns); err != nil {
		return err
	}

	// Wait for the PV's phase to return to be `expectPVPhase`
	Logf("Waiting for reclaim process to complete.")
	if err = WaitForPersistentVolumePhase(expectPVPhase, c, pv.Name, 1*time.Second, 300*time.Second); err != nil {
		return fmt.Errorf("pv %q phase did not become %v: %v", pv.Name, expectPVPhase, err)
	}

	// examine the pv's ClaimRef and UID and compare to expected values
	if pv, err = c.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("PV Get API error: %v", err)
	}
	cr := pv.Spec.ClaimRef
	if expectPVPhase == v1.VolumeAvailable {
		if cr != nil && len(cr.UID) > 0 {
			return fmt.Errorf("PV is 'Available' but ClaimRef.UID is not empty")
		}
	} else if expectPVPhase == v1.VolumeBound {
		if cr == nil {
			return fmt.Errorf("PV is 'Bound' but ClaimRef is nil")
		}
		if len(cr.UID) == 0 {
			return fmt.Errorf("PV is 'Bound' but ClaimRef.UID is empty")
		}
	}

	Logf("PV %v now in %q phase", pv.Name, expectPVPhase)
	return err
}

// Wraps deletePVCandValidatePV() by calling the function in a loop over the PV map. Only bound PVs
// are deleted. Validates that the claim was deleted and the PV is in the expected Phase (Released,
// Available, Bound).
// Note: if there are more claims than pvs then some of the remaining claims will bind to the
//   just-made-available pvs.
func DeletePVCandValidatePVGroup(c clientset.Interface, ns string, pvols PVMap, claims PVCMap, expectPVPhase v1.PersistentVolumePhase) error {

	var boundPVs, deletedPVCs int

	for pvName := range pvols {
		pv, err := c.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("PV Get API error: %v", err)
		}
		cr := pv.Spec.ClaimRef
		// if pv is bound then delete the pvc it is bound to
		if cr != nil && len(cr.Name) > 0 {
			boundPVs++
			// Assert bound PVC is tracked in this test. Failing this might
			// indicate external PVCs interfering with the test.
			pvcKey := makePvcKey(ns, cr.Name)
			if _, found := claims[pvcKey]; !found {
				return fmt.Errorf("internal: claims map is missing pvc %q", pvcKey)
			}
			pvc, err := c.CoreV1().PersistentVolumeClaims(ns).Get(cr.Name, metav1.GetOptions{})
			if err != nil && !apierrs.IsNotFound(err) {
				return fmt.Errorf("PVC Get API error: %v", err)
			}
			if err = DeletePVCandValidatePV(c, ns, pvc, pv, expectPVPhase); err != nil {
				return err
			}
			delete(claims, pvcKey)
			deletedPVCs++
		}
	}
	if boundPVs != deletedPVCs {
		return fmt.Errorf("Expect number of bound PVs (%v) to equal number of deleted PVCs (%v)", boundPVs, deletedPVCs)
	}
	return nil
}

// create the PV resource. Fails test on error.
func createPV(c clientset.Interface, pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	pv, err := c.CoreV1().PersistentVolumes().Create(pv)
	if err != nil {
		return nil, fmt.Errorf("PV Create API error: %v", err)
	}
	return pv, nil
}

// create the PVC resource. Fails test on error.
func CreatePVC(c clientset.Interface, ns string, pvc *v1.PersistentVolumeClaim) (*v1.PersistentVolumeClaim, error) {
	pvc, err := c.CoreV1().PersistentVolumeClaims(ns).Create(pvc)
	if err != nil {
		return nil, fmt.Errorf("PVC Create API error: %v", err)
	}
	return pvc, nil
}

// Create a PVC followed by the PV based on the passed in nfs-server ip and
// namespace. If the "preBind" bool is true then pre-bind the PV to the PVC
// via the PV's ClaimRef. Return the pv and pvc to reflect the created objects.
// Note: in the pre-bind case the real PVC name, which is generated, is not
//   known until after the PVC is instantiated. This is why the pvc is created
//   before the pv.
func CreatePVCPV(c clientset.Interface, pvConfig PersistentVolumeConfig, pvcConfig PersistentVolumeClaimConfig, ns string, preBind bool) (*v1.PersistentVolume, *v1.PersistentVolumeClaim, error) {

	var preBindMsg string
	var err error

	// make the pvc definition first
	pvc := MakePersistentVolumeClaim(pvcConfig, ns)
	if preBind {
		preBindMsg = " pre-bound"
		pvConfig.Prebind = pvc
	}
	// make the pv spec
	pv := MakePersistentVolume(pvConfig)

	By(fmt.Sprintf("Creating a PVC followed by a%s PV", preBindMsg))
	// instantiate the pvc
	if pvc, err = CreatePVC(c, ns, pvc); err != nil {
		return nil, nil, err
	}

	// instantiate the pv, handle pre-binding by ClaimRef if needed
	if preBind {
		pv.Spec.ClaimRef.Name = pvc.Name
	}
	if pv, err = createPV(c, pv); err != nil {
		return nil, pvc, err
	}
	return pv, pvc, nil
}

// Create a PV followed by the PVC based on the passed in nfs-server ip and
// namespace. If the "preBind" bool is true then pre-bind the PVC to the PV
// via the PVC's VolumeName. Return the pv and pvc to reflect the created
// objects.
// Note: in the pre-bind case the real PV name, which is generated, is not
//   known until after the PV is instantiated. This is why the pv is created
//   before the pvc.
func CreatePVPVC(c clientset.Interface, pvConfig PersistentVolumeConfig, pvcConfig PersistentVolumeClaimConfig, ns string, preBind bool) (*v1.PersistentVolume, *v1.PersistentVolumeClaim, error) {

	var err error
	preBindMsg := ""
	if preBind {
		preBindMsg = " pre-bound"
	}
	Logf("Creating a PV followed by a%s PVC", preBindMsg)

	// make the pv and pvc definitions
	pv := MakePersistentVolume(pvConfig)
	pvc := MakePersistentVolumeClaim(pvcConfig, ns)

	// instantiate the pv
	if pv, err = createPV(c, pv); err != nil {
		return nil, nil, err
	}
	// instantiate the pvc, handle pre-binding by VolumeName if needed
	if preBind {
		pvc.Spec.VolumeName = pv.Name
	}
	if pvc, err = CreatePVC(c, ns, pvc); err != nil {
		return pv, nil, err
	}
	return pv, pvc, nil
}

// Create the desired number of PVs and PVCs and return them in separate maps. If the
// number of PVs != the number of PVCs then the min of those two counts is the number of
// PVs expected to bind.
func CreatePVsPVCs(numpvs, numpvcs int, c clientset.Interface, ns string, pvConfig PersistentVolumeConfig, pvcConfig PersistentVolumeClaimConfig) (PVMap, PVCMap, error) {

	var (
		i, extraPVs, extraPVCs int
		pv                     *v1.PersistentVolume
		pvc                    *v1.PersistentVolumeClaim
		err                    error
	)
	pvMap := make(PVMap, numpvs)
	pvcMap := make(PVCMap, numpvcs)

	extraPVs = numpvs - numpvcs
	if extraPVs < 0 {
		extraPVCs = -extraPVs
		extraPVs = 0
	}
	pvsToCreate := numpvs - extraPVs // want the min(numpvs, numpvcs)

	// create pvs and pvcs
	for i = 0; i < pvsToCreate; i++ {
		if pv, pvc, err = CreatePVPVC(c, pvConfig, pvcConfig, ns, false); err != nil {
			return pvMap, pvcMap, err
		}
		pvMap[pv.Name] = pvval{}
		pvcMap[makePvcKey(ns, pvc.Name)] = pvcval{}
	}

	// create extra pvs or pvcs as needed
	for i = 0; i < extraPVs; i++ {
		pv = MakePersistentVolume(pvConfig)
		if pv, err = createPV(c, pv); err != nil {
			return pvMap, pvcMap, err
		}
		pvMap[pv.Name] = pvval{}
	}
	for i = 0; i < extraPVCs; i++ {
		pvc = MakePersistentVolumeClaim(pvcConfig, ns)
		if pvc, err = CreatePVC(c, ns, pvc); err != nil {
			return pvMap, pvcMap, err
		}
		pvcMap[makePvcKey(ns, pvc.Name)] = pvcval{}
	}
	return pvMap, pvcMap, nil
}

// Wait for the pv and pvc to bind to each other.
func WaitOnPVandPVC(c clientset.Interface, ns string, pv *v1.PersistentVolume, pvc *v1.PersistentVolumeClaim) error {

	var err error

	// Wait for newly created PVC to bind to the PV
	Logf("Waiting for PV %v to bind to PVC %v", pv.Name, pvc.Name)
	if err = WaitForPersistentVolumeClaimPhase(v1.ClaimBound, c, ns, pvc.Name, 3*time.Second, 300*time.Second); err != nil {
		return fmt.Errorf("PVC %q did not become Bound: %v", pvc.Name, err)
	}

	// Wait for PersistentVolume.Status.Phase to be Bound, which it should be
	// since the PVC is already bound.
	if err = WaitForPersistentVolumePhase(v1.VolumeBound, c, pv.Name, 3*time.Second, 300*time.Second); err != nil {
		return fmt.Errorf("PV %q did not become Bound: %v", pv.Name, err)
	}

	// Re-get the pv and pvc objects
	if pv, err = c.CoreV1().PersistentVolumes().Get(pv.Name, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("PV Get API error: %v", err)
	}
	// Re-get the pvc and
	if pvc, err = c.CoreV1().PersistentVolumeClaims(ns).Get(pvc.Name, metav1.GetOptions{}); err != nil {
		return fmt.Errorf("PVC Get API error: %v", err)
	}

	// The pv and pvc are both bound, but to each other?
	// Check that the PersistentVolume.ClaimRef matches the PVC
	if pv.Spec.ClaimRef == nil {
		return fmt.Errorf("PV %q ClaimRef is nil", pv.Name)
	}
	if pv.Spec.ClaimRef.Name != pvc.Name {
		return fmt.Errorf("PV %q ClaimRef's name (%q) should be %q", pv.Name, pv.Spec.ClaimRef.Name, pvc.Name)
	}
	if pvc.Spec.VolumeName != pv.Name {
		return fmt.Errorf("PVC %q VolumeName (%q) should be %q", pvc.Name, pvc.Spec.VolumeName, pv.Name)
	}
	if pv.Spec.ClaimRef.UID != pvc.UID {
		return fmt.Errorf("PV %q ClaimRef's UID (%q) should be %q", pv.Name, pv.Spec.ClaimRef.UID, pvc.UID)
	}

	return nil
}

// Search for bound PVs and PVCs by examining pvols for non-nil claimRefs.
// NOTE: Each iteration waits for a maximum of 3 minutes per PV and, if the PV is bound,
//   up to 3 minutes for the PVC. When the number of PVs != number of PVCs, this can lead
//   to situations where the maximum wait times are reached several times in succession,
//   extending test time. Thus, it is recommended to keep the delta between PVs and PVCs
//   small.
func WaitAndVerifyBinds(c clientset.Interface, ns string, pvols PVMap, claims PVCMap, testExpected bool) error {

	var err error
	var actualBinds int
	expectedBinds := len(pvols)
	if expectedBinds > len(claims) { // want the min of # pvs or #pvcs
		expectedBinds = len(claims)
	}

	for pvName := range pvols {
		err = WaitForPersistentVolumePhase(v1.VolumeBound, c, pvName, 3*time.Second, 180*time.Second)
		if err != nil && len(pvols) > len(claims) {
			Logf("WARN: pv %v is not bound after max wait", pvName)
			Logf("      This may be ok since there are more pvs than pvcs")
			continue
		}
		if err != nil {
			return fmt.Errorf("PV %q did not become Bound: %v", pvName, err)
		}

		pv, err := c.CoreV1().PersistentVolumes().Get(pvName, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("PV Get API error: %v", err)
		}
		cr := pv.Spec.ClaimRef
		if cr != nil && len(cr.Name) > 0 {
			// Assert bound pvc is a test resource. Failing assertion could
			// indicate non-test PVC interference or a bug in the test
			pvcKey := makePvcKey(ns, cr.Name)
			if _, found := claims[pvcKey]; !found {
				return fmt.Errorf("internal: claims map is missing pvc %q", pvcKey)
			}

			if err = WaitForPersistentVolumeClaimPhase(v1.ClaimBound, c, ns, cr.Name, 3*time.Second, 180*time.Second); err != nil {
				return fmt.Errorf("PVC %q did not become Bound: %v", cr.Name, err)
			}
			actualBinds++
		}
	}

	if testExpected && actualBinds != expectedBinds {
		return fmt.Errorf("Expect number of bound PVs (%v) to equal number of claims (%v)", actualBinds, expectedBinds)
	}
	return nil
}

// Test the pod's exit code to be zero.
func testPodSuccessOrFail(c clientset.Interface, ns string, pod *v1.Pod) error {

	By("Pod should terminate with exitcode 0 (success)")
	if err := WaitForPodSuccessInNamespace(c, pod.Name, ns); err != nil {
		fmt.Errorf("pod %q failed to reach Success: %v", pod.Name, err)
	}
	Logf("Pod %v succeeded ", pod.Name)
	return nil
}

// Deletes the passed-in pod and waits for the pod to be terminated. Resilient to the pod
// not existing.
func DeletePodWithWait(f *Framework, c clientset.Interface, pod *v1.Pod) error {

	if pod == nil {
		return nil
	}
	Logf("Deleting pod %v", pod.Name)
	err := c.CoreV1().Pods(pod.Namespace).Delete(pod.Name, nil)
	if err != nil {
		if apierrs.IsNotFound(err) {
			return nil // assume pod was deleted already
		}
		return fmt.Errorf("Pod Get API error: %v", err)
	}

	// wait for pod to terminate. Expect apierr NotFound
	err = f.WaitForPodTerminated(pod.Name, "")
	if err != nil && !apierrs.IsNotFound(err) {
		return fmt.Errorf("error deleting pod %q: %v", pod.Name, err)
	}
	if apierrs.IsNotFound(err) {
		Logf("Ignore \"not found\" error above. Pod %q successfully deleted", pod.Name)
	}
	return nil
}

// Create the test pod, wait for (hopefully) success, and then delete the pod.
// Note: need named return value so that the err assignment in the defer sets the returned error.
func CreateWaitAndDeletePod(f *Framework, c clientset.Interface, ns string, pvc *v1.PersistentVolumeClaim) (err error) {

	Logf("Creating nfs test pod")
	// Make pod spec
	pod := MakeWritePod(ns, pvc)

	// Instantiate pod (Create)
	runPod, err := c.CoreV1().Pods(ns).Create(pod)
	if err != nil {
		return fmt.Errorf("Pod Create API error: %v", err)
	}
	defer func() {
		delErr := DeletePodWithWait(f, c, runPod)
		if err == nil { // don't override previous err value
			err = delErr // assign to returned err, can be nil
		}
	}()

	// Wait for the test pod to complete its lifecycle
	err = testPodSuccessOrFail(c, ns, runPod)
	if err != nil {
		return fmt.Errorf("Pod %q did not exit with Success: %v", runPod.Name, err)
	}
	return // named return value
}

// Sanity check for GCE testing.  Verify the persistent disk attached to the node.
func VerifyGCEDiskAttached(diskName string, nodeName types.NodeName) (bool, error) {

	var err error
	var isAttached bool
	gceCloud, err := GetGCECloud()
	if err != nil {
		return false, fmt.Errorf("GetGCECloud error: %v", err)
	}
	if isAttached, err = gceCloud.DiskIsAttached(diskName, nodeName); err != nil {
		return false, fmt.Errorf("Cannot verify if GCE disk is attached: %v", err)
	}
	return isAttached, nil
}

// Return a pvckey struct.
func makePvcKey(ns, name string) types.NamespacedName {
	return types.NamespacedName{Namespace: ns, Name: name}
}

// Returns a PV definition based on the nfs server IP. If the PVC is not nil
// then the PV is defined with a ClaimRef which includes the PVC's namespace.
// If the PVC is nil then the PV is not defined with a ClaimRef.  If no reclaimPolicy
// is assigned, assumes "Retain".
// Note: the passed-in claim does not have a name until it is created
//   (instantiated) and thus the PV's ClaimRef cannot be completely filled-in in
//   this func. Therefore, the ClaimRef's name is added later in
//   createPVCPV.
func MakePersistentVolume(pvConfig PersistentVolumeConfig) *v1.PersistentVolume {
	// Specs are expected to match the test's PersistentVolumeClaim

	var claimRef *v1.ObjectReference
	// If the reclaimPolicy is not provided, assume Retain
	if pvConfig.ReclaimPolicy == "" {
		Logf("PV ReclaimPolicy unspecified, default: Retain")
		pvConfig.ReclaimPolicy = v1.PersistentVolumeReclaimRetain
	}
	if pvConfig.Prebind != nil {
		claimRef = &v1.ObjectReference{
			Name:      pvConfig.Prebind.Name,
			Namespace: pvConfig.Prebind.Namespace,
		}
	}
	return &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: pvConfig.NamePrefix,
			Labels:       pvConfig.Labels,
			Annotations: map[string]string{
				volumehelper.VolumeGidAnnotationKey: "777",
			},
		},
		Spec: v1.PersistentVolumeSpec{
			PersistentVolumeReclaimPolicy: pvConfig.ReclaimPolicy,
			Capacity: v1.ResourceList{
				v1.ResourceName(v1.ResourceStorage): resource.MustParse("2Gi"),
			},
			PersistentVolumeSource: pvConfig.PVSource,
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
				v1.ReadOnlyMany,
				v1.ReadWriteMany,
			},
			ClaimRef: claimRef,
		},
	}
}

// Returns a PVC definition based on the namespace.
// Note: if this PVC is intended to be pre-bound to a PV, whose name is not
//   known until the PV is instantiated, then the func CreatePVPVC will add
//   pvc.Spec.VolumeName to this claim.
func MakePersistentVolumeClaim(cfg PersistentVolumeClaimConfig, ns string) *v1.PersistentVolumeClaim {
	// Specs are expected to match this test's PersistentVolume

	if len(cfg.AccessModes) == 0 {
		Logf("AccessModes unspecified, default: all modes (RWO, RWX, ROX).")
		cfg.AccessModes = append(cfg.AccessModes, v1.ReadWriteOnce, v1.ReadOnlyMany, v1.ReadOnlyMany)
	}

	return &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-",
			Namespace:    ns,
			Annotations:  cfg.Annotations,
		},
		Spec: v1.PersistentVolumeClaimSpec{
			Selector:    cfg.Selector,
			AccessModes: cfg.AccessModes,
			Resources: v1.ResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceName(v1.ResourceStorage): resource.MustParse("1Gi"),
				},
			},
		},
	}
}

func CreatePDWithRetry() (string, error) {
	newDiskName := ""
	var err error
	for start := time.Now(); time.Since(start) < PDRetryTimeout; time.Sleep(PDRetryPollTime) {
		if newDiskName, err = createPD(); err != nil {
			Logf("Couldn't create a new PD. Sleeping 5 seconds (%v)", err)
			continue
		}
		Logf("Successfully created a new PD: %q.", newDiskName)
		break
	}
	return newDiskName, err
}

func DeletePDWithRetry(diskName string) error {
	var err error
	for start := time.Now(); time.Since(start) < PDRetryTimeout; time.Sleep(PDRetryPollTime) {
		if err = deletePD(diskName); err != nil {
			Logf("Couldn't delete PD %q. Sleeping %v: %v", diskName, PDRetryPollTime, err)
			continue
		}
		Logf("Successfully deleted PD %q.", diskName)
		break
	}
	if err != nil {
		return fmt.Errorf("unable to delete PD %q: %v", diskName, err)
	}
	return nil
}

func createPD() (string, error) {
	if TestContext.Provider == "gce" || TestContext.Provider == "gke" {
		pdName := fmt.Sprintf("%s-%s", TestContext.Prefix, string(uuid.NewUUID()))

		gceCloud, err := GetGCECloud()
		if err != nil {
			return "", err
		}

		tags := map[string]string{}
		err = gceCloud.CreateDisk(pdName, gcecloud.DiskTypeSSD, TestContext.CloudConfig.Zone, 10 /* sizeGb */, tags)
		if err != nil {
			return "", err
		}
		return pdName, nil
	} else if TestContext.Provider == "aws" {
		client := ec2.New(session.New())

		request := &ec2.CreateVolumeInput{}
		request.AvailabilityZone = aws.String(TestContext.CloudConfig.Zone)
		request.Size = aws.Int64(10)
		request.VolumeType = aws.String(awscloud.DefaultVolumeType)
		response, err := client.CreateVolume(request)
		if err != nil {
			return "", err
		}

		az := aws.StringValue(response.AvailabilityZone)
		awsID := aws.StringValue(response.VolumeId)

		volumeName := "aws://" + az + "/" + awsID
		return volumeName, nil
	} else {
		return "", fmt.Errorf("Provider does not support volume creation")
	}
}

func deletePD(pdName string) error {
	if TestContext.Provider == "gce" || TestContext.Provider == "gke" {
		gceCloud, err := GetGCECloud()
		if err != nil {
			return err
		}

		err = gceCloud.DeleteDisk(pdName)

		if err != nil {
			if gerr, ok := err.(*googleapi.Error); ok && len(gerr.Errors) > 0 && gerr.Errors[0].Reason == "notFound" {
				// PD already exists, ignore error.
				return nil
			}

			Logf("Error deleting PD %q: %v", pdName, err)
		}
		return err
	} else if TestContext.Provider == "aws" {
		client := ec2.New(session.New())

		tokens := strings.Split(pdName, "/")
		awsVolumeID := tokens[len(tokens)-1]

		request := &ec2.DeleteVolumeInput{VolumeId: aws.String(awsVolumeID)}
		_, err := client.DeleteVolume(request)
		if err != nil {
			if awsError, ok := err.(awserr.Error); ok && awsError.Code() == "InvalidVolume.NotFound" {
				Logf("Volume deletion implicitly succeeded because volume %q does not exist.", pdName)
			} else {
				return fmt.Errorf("error deleting EBS volumes: %v", err)
			}
		}
		return nil
	} else {
		return fmt.Errorf("Provider does not support volume deletion")
	}
}

// Returns a pod definition based on the namespace. The pod references the PVC's
// name.
func MakeWritePod(ns string, pvc *v1.PersistentVolumeClaim) *v1.Pod {
	return MakePod(ns, []*v1.PersistentVolumeClaim{pvc}, true, "touch /mnt/volume1/SUCCESS && (id -G | grep -E '\\b777\\b')")
}

// Returns a pod definition based on the namespace. The pod references the PVC's
// name.  A slice of BASH commands can be supplied as args to be run by the pod
func MakePod(ns string, pvclaims []*v1.PersistentVolumeClaim, isPrivileged bool, command string) *v1.Pod {
	if len(command) == 0 {
		command = "while true; do sleep 1; done"
	}
	podSpec := &v1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Pod",
			APIVersion: api.Registry.GroupOrDie(v1.GroupName).GroupVersion.String(),
		},
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "pvc-tester-",
			Namespace:    ns,
		},
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{
					Name:    "write-pod",
					Image:   "gcr.io/google_containers/busybox:1.24",
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", command},
					SecurityContext: &v1.SecurityContext{
						Privileged: &isPrivileged,
					},
				},
			},
			RestartPolicy: v1.RestartPolicyOnFailure,
		},
	}
	var volumeMounts = make([]v1.VolumeMount, len(pvclaims))
	var volumes = make([]v1.Volume, len(pvclaims))
	for index, pvclaim := range pvclaims {
		volumename := fmt.Sprintf("volume%v", index+1)
		volumeMounts[index] = v1.VolumeMount{Name: volumename, MountPath: "/mnt/" + volumename}
		volumes[index] = v1.Volume{Name: volumename, VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: pvclaim.Name, ReadOnly: false}}}
	}
	podSpec.Spec.Containers[0].VolumeMounts = volumeMounts
	podSpec.Spec.Volumes = volumes
	return podSpec
}

// create pod with given claims
func CreatePod(client clientset.Interface, namespace string, pvclaims []*v1.PersistentVolumeClaim, isPrivileged bool, command string) (*v1.Pod, error) {

	var err error
	pod := MakePod(namespace, pvclaims, isPrivileged, command)
	if pod, err = client.CoreV1().Pods(namespace).Create(pod); err != nil {
		return nil, fmt.Errorf("Pod Create API error: %v", err)
	}
	// Waiting for pod to be running
	if err = WaitForPodNameRunningInNamespace(client, pod.Name, namespace); err != nil {
		return pod, fmt.Errorf("pod %q is not Running: %v", pod.Name, err)
	}
	// get fresh pod info
	if pod, err = client.CoreV1().Pods(namespace).Get(pod.Name, metav1.GetOptions{}); err != nil {
		return pod, fmt.Errorf("Pod Get API error: %v", err)
	}
	return pod, nil
}

// Define and create a pod with a mounted PV.  Pod runs infinite loop until killed.
func CreateClientPod(c clientset.Interface, ns string, pvc *v1.PersistentVolumeClaim) (*v1.Pod, error) {
	return CreatePod(c, ns, []*v1.PersistentVolumeClaim{pvc}, true, "")
}

// wait until all pvcs phase set to bound
func WaitForPVClaimBoundPhase(client clientset.Interface, pvclaims []*v1.PersistentVolumeClaim) ([]*v1.PersistentVolume, error) {

	var err error
	var persistentvolumes = make([]*v1.PersistentVolume, len(pvclaims))

	for index, claim := range pvclaims {
		if err = WaitForPersistentVolumeClaimPhase(v1.ClaimBound, client, claim.Namespace, claim.Name, Poll, ClaimProvisionTimeout); err != nil {
			return persistentvolumes, err
		}
		// Get new copy of the claim
		if claim, err = client.CoreV1().PersistentVolumeClaims(claim.Namespace).Get(claim.Name, metav1.GetOptions{}); err != nil {
			return persistentvolumes, fmt.Errorf("PVC Get API error: %v", err)
		}
		// Get the bounded PV
		if persistentvolumes[index], err = client.CoreV1().PersistentVolumes().Get(claim.Spec.VolumeName, metav1.GetOptions{}); err != nil {
			return persistentvolumes, fmt.Errorf("PV Get API error: %v", err)
		}
	}
	return persistentvolumes, nil
}
