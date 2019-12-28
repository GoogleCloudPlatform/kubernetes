/*
Copyright 2018 The Kubernetes Authors.

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

package state

import (
	"fmt"
	"os"
	"path"
	"sync"

	"k8s.io/klog"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager"
	"k8s.io/kubernetes/pkg/kubelet/checkpointmanager/errors"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpumanager/containermap"
	"k8s.io/kubernetes/pkg/kubelet/cm/cpuset"
)

var _ State = &stateCheckpoint{}

type stateCheckpoint struct {
	mux               sync.RWMutex
	policyName        string
	cache             State
	checkpointManager checkpointmanager.CheckpointManager
	checkpointName    string
	initialContainers containermap.ContainerMap
}

// NewCheckpointState creates new State for keeping track of cpu/pod assignment with checkpoint backend
func NewCheckpointState(stateDir, checkpointName, policyName string, initialContainers containermap.ContainerMap) (State, error) {
	checkpointManager, err := checkpointmanager.NewCheckpointManager(stateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize checkpoint manager: %v", err)
	}
	stateCheckpoint := &stateCheckpoint{
		cache:             NewMemoryState(),
		policyName:        policyName,
		checkpointManager: checkpointManager,
		checkpointName:    checkpointName,
		initialContainers: initialContainers,
	}

	if err := stateCheckpoint.restoreState(); err != nil {
		klog.Errorf("could not restore state from checkpoint: %v", err)
		checkpointFile := path.Join(stateDir, checkpointName)
		corruptpath := checkpointFile + ".corrupt"
		if err := os.Rename(checkpointFile, corruptpath); err != nil {
			klog.Warningf("cannot move %s : %v", checkpointFile, err)
			if err := os.Remove(checkpointFile); err != nil {
				//lint:ignore ST1005 user-facing error message
				return nil, fmt.Errorf("could not delete checkpoint: %v\n"+
					"Please drain this node and delete the CPU manager checkpoint file %q before restarting Kubelet.",
					err, checkpointFile)
			}
			klog.Infof("deleted %s", checkpointFile)
		} else {
			klog.Infof("moved %s to %s", checkpointFile, corruptpath)
		}
	}

	return stateCheckpoint, nil
}

// migrateV1CheckpointToV2Checkpoint() converts checkpoints from the v1 format to the v2 format
func (sc *stateCheckpoint) migrateV1CheckpointToV2Checkpoint(src *CPUManagerCheckpointV1, dst *CPUManagerCheckpointV2) error {
	if src.PolicyName != "" {
		dst.PolicyName = src.PolicyName
	}
	if src.DefaultCPUSet != "" {
		dst.DefaultCPUSet = src.DefaultCPUSet
	}
	for containerID, cset := range src.Entries {
		podUID, containerName, err := sc.initialContainers.GetContainerRef(containerID)
		if err != nil {
			return fmt.Errorf("containerID '%v' not found in initial containers list", containerID)
		}
		if dst.Entries == nil {
			dst.Entries = make(map[string]map[string]string)
		}
		if _, exists := dst.Entries[podUID]; !exists {
			dst.Entries[podUID] = make(map[string]string)
		}
		dst.Entries[podUID][containerName] = cset
	}
	return nil
}

// restores state from a checkpoint and creates it if it doesn't exist
func (sc *stateCheckpoint) restoreState() error {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	var err error

	// used when all parsing is ok
	tmpAssignments := make(ContainerCPUAssignments)
	tmpDefaultCPUSet := cpuset.NewCPUSet()
	tmpContainerCPUSet := cpuset.NewCPUSet()

	checkpointV1 := newCPUManagerCheckpointV1()
	checkpointV2 := newCPUManagerCheckpointV2()

	if err = sc.checkpointManager.GetCheckpoint(sc.checkpointName, checkpointV1); err != nil {
		checkpointV1 = &CPUManagerCheckpointV1{} // reset it back to 0
		if err = sc.checkpointManager.GetCheckpoint(sc.checkpointName, checkpointV2); err != nil {
			if err == errors.ErrCheckpointNotFound {
				sc.storeState()
				return nil
			}
			return err
		}
	}

	if err = sc.migrateV1CheckpointToV2Checkpoint(checkpointV1, checkpointV2); err != nil {
		return fmt.Errorf("error migrating v1 checkpoint state to v2 checkpoint state: %s", err)
	}

	if sc.policyName != checkpointV2.PolicyName {
		return fmt.Errorf("configured policy %q differs from state checkpoint policy %q", sc.policyName, checkpointV2.PolicyName)
	}

	if tmpDefaultCPUSet, err = cpuset.Parse(checkpointV2.DefaultCPUSet); err != nil {
		return fmt.Errorf("could not parse default cpu set %q: %v", checkpointV2.DefaultCPUSet, err)
	}

	for pod := range checkpointV2.Entries {
		tmpAssignments[pod] = make(map[string]cpuset.CPUSet)
		for container, cpuString := range checkpointV2.Entries[pod] {
			if tmpContainerCPUSet, err = cpuset.Parse(cpuString); err != nil {
				return fmt.Errorf("could not parse cpuset %q for container %q in pod %q: %v", cpuString, container, pod, err)
			}
			tmpAssignments[pod][container] = tmpContainerCPUSet
		}
	}

	sc.cache.SetDefaultCPUSet(tmpDefaultCPUSet)
	sc.cache.SetCPUAssignments(tmpAssignments)

	klog.V(2).Info("[cpumanager] state checkpoint: restored state from checkpoint")
	klog.V(2).Infof("[cpumanager] state checkpoint: defaultCPUSet: %s", tmpDefaultCPUSet.String())

	return nil
}

// saves state to a checkpoint, caller is responsible for locking
func (sc *stateCheckpoint) storeState() {
	checkpoint := NewCPUManagerCheckpoint()
	checkpoint.PolicyName = sc.policyName
	checkpoint.DefaultCPUSet = sc.cache.GetDefaultCPUSet().String()

	assignments := sc.cache.GetCPUAssignments()
	for pod := range assignments {
		checkpoint.Entries[pod] = make(map[string]string)
		for container, cset := range assignments[pod] {
			checkpoint.Entries[pod][container] = cset.String()
		}
	}

	err := sc.checkpointManager.CreateCheckpoint(sc.checkpointName, checkpoint)

	if err != nil {
		panic("[cpumanager] could not save checkpoint: " + err.Error())
	}
}

// GetCPUSet returns current CPU set
func (sc *stateCheckpoint) GetCPUSet(podUID string, containerName string) (cpuset.CPUSet, bool) {
	sc.mux.RLock()
	defer sc.mux.RUnlock()

	res, ok := sc.cache.GetCPUSet(podUID, containerName)
	return res, ok
}

// GetDefaultCPUSet returns default CPU set
func (sc *stateCheckpoint) GetDefaultCPUSet() cpuset.CPUSet {
	sc.mux.RLock()
	defer sc.mux.RUnlock()

	return sc.cache.GetDefaultCPUSet()
}

// GetCPUSetOrDefault returns current CPU set, or default one if it wasn't changed
func (sc *stateCheckpoint) GetCPUSetOrDefault(podUID string, containerName string) cpuset.CPUSet {
	sc.mux.RLock()
	defer sc.mux.RUnlock()

	return sc.cache.GetCPUSetOrDefault(podUID, containerName)
}

// GetCPUAssignments returns current CPU to pod assignments
func (sc *stateCheckpoint) GetCPUAssignments() ContainerCPUAssignments {
	sc.mux.RLock()
	defer sc.mux.RUnlock()

	return sc.cache.GetCPUAssignments()
}

// SetCPUSet sets CPU set
func (sc *stateCheckpoint) SetCPUSet(podUID string, containerName string, cset cpuset.CPUSet) {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	sc.cache.SetCPUSet(podUID, containerName, cset)
	sc.storeState()
}

// SetDefaultCPUSet sets default CPU set
func (sc *stateCheckpoint) SetDefaultCPUSet(cset cpuset.CPUSet) {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	sc.cache.SetDefaultCPUSet(cset)
	sc.storeState()
}

// SetCPUAssignments sets CPU to pod assignments
func (sc *stateCheckpoint) SetCPUAssignments(a ContainerCPUAssignments) {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	sc.cache.SetCPUAssignments(a)
	sc.storeState()
}

// Delete deletes assignment for specified pod
func (sc *stateCheckpoint) Delete(podUID string, containerName string) {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	sc.cache.Delete(podUID, containerName)
	sc.storeState()
}

// ClearState clears the state and saves it in a checkpoint
func (sc *stateCheckpoint) ClearState() {
	sc.mux.Lock()
	defer sc.mux.Unlock()
	sc.cache.ClearState()
	sc.storeState()
}
