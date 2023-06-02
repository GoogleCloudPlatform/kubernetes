/*
Copyright 2023 The Kubernetes Authors.

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

package job

import (
	"fmt"
	"sort"
	"time"

	v1 "k8s.io/api/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/utils/clock"
	"k8s.io/utils/pointer"

	apipod "k8s.io/kubernetes/pkg/api/v1/pod"
)

type backoffRecord struct {
	key                      string
	failuresAfterLastSuccess int32
	lastFailureTime          *time.Time
}

type backoffStore struct {
	store cache.Store
}

func (s *backoffStore) updateBackoffRecord(record backoffRecord) error {
	b, ok, err := s.store.GetByKey(record.key)
	if err != nil {
		return err
	}

	if !ok {
		err = s.store.Add(&record)
		if err != nil {
			return err
		}
	} else {
		backoffRecord := b.(*backoffRecord)
		backoffRecord.failuresAfterLastSuccess = record.failuresAfterLastSuccess
		backoffRecord.lastFailureTime = record.lastFailureTime
	}

	return nil
}

func (s *backoffStore) removeBackoffRecord(jobId string) error {
	b, ok, err := s.store.GetByKey(jobId)
	if err != nil {
		return err
	}

	if ok {
		err = s.store.Delete(b)
		if err != nil {
			return err
		}
	}

	return nil

}

func newBackoffRecordStore() *backoffStore {
	return &backoffStore{
		store: cache.NewStore(backoffRecordKeyFunc),
	}
}

var backoffRecordKeyFunc = func(obj interface{}) (string, error) {
	if u, ok := obj.(*backoffRecord); ok {
		return u.key, nil
	}
	return "", fmt.Errorf("could not find key for obj %#v", obj)
}

func (backoffRecordStore *backoffStore) newBackoffRecord(key string, newSucceededPods []*v1.Pod, newFailedPods []*v1.Pod) backoffRecord {
	var backoff *backoffRecord

	if b, exists, _ := backoffRecordStore.store.GetByKey(key); exists {
		old := b.(*backoffRecord)
		backoff = &backoffRecord{
			key:                      old.key,
			failuresAfterLastSuccess: old.failuresAfterLastSuccess,
			lastFailureTime:          old.lastFailureTime,
		}
	} else {
		backoff = &backoffRecord{
			key:                      key,
			failuresAfterLastSuccess: 0,
			lastFailureTime:          nil,
		}
	}

	sortByFinishedTime(newSucceededPods)
	sortByFinishedTime(newFailedPods)

	if len(newSucceededPods) == 0 {
		if len(newFailedPods) == 0 {
			return *backoff
		}

		backoff.failuresAfterLastSuccess = backoff.failuresAfterLastSuccess + int32(len(newFailedPods))
		lastFailureTime := getFinishedTime(newFailedPods[len(newFailedPods)-1])
		backoff.lastFailureTime = &lastFailureTime
		return *backoff

	} else {
		if len(newFailedPods) == 0 {
			backoff.failuresAfterLastSuccess = 0
			backoff.lastFailureTime = nil
			return *backoff
		}

		backoff.failuresAfterLastSuccess = 0
		backoff.lastFailureTime = nil

		lastSuccessTime := getFinishedTime(newSucceededPods[len(newSucceededPods)-1])
		for i := len(newFailedPods) - 1; i >= 0; i-- {
			failedTime := getFinishedTime(newFailedPods[i])
			if !failedTime.After(lastSuccessTime) {
				break
			}
			if backoff.lastFailureTime == nil {
				backoff.lastFailureTime = &failedTime
			}
			backoff.failuresAfterLastSuccess += 1
		}

		return *backoff

	}

}

func sortByFinishedTime(pods []*v1.Pod) {
	sort.Slice(pods, func(i, j int) bool {
		p1 := pods[i]
		p2 := pods[j]
		p1FinishTime := getFinishedTime(p1)
		p2FinishTime := getFinishedTime(p2)

		return p1FinishTime.Before(p2FinishTime)
	})
}

func getFinishedTime(p *v1.Pod) time.Time {
	finishTime := getFinishTimeFromContainers(p)
	if finishTime == nil {
		finishTime = getFinishTimeFromPodReadyFalseCondition(p)
	}
	if finishTime == nil {
		finishTime = getFinishTimeFromDeletionTimestamp(p)
	}
	if finishTime != nil {
		return *finishTime
	}
	return p.CreationTimestamp.Time
}

func getFinishTimeFromContainers(p *v1.Pod) *time.Time {
	var finishTime *time.Time
	for _, containerState := range p.Status.ContainerStatuses {
		if containerState.State.Terminated == nil {
			return nil
		}
		if containerState.State.Terminated.FinishedAt.Time.IsZero() {
			return nil
		}
		if finishTime == nil || finishTime.Before(containerState.State.Terminated.FinishedAt.Time) {
			finishTime = &containerState.State.Terminated.FinishedAt.Time
		}
	}
	return finishTime
}

func getFinishTimeFromPodReadyFalseCondition(p *v1.Pod) *time.Time {
	if _, c := apipod.GetPodCondition(&p.Status, v1.PodReady); c != nil && c.Status == v1.ConditionFalse && !c.LastTransitionTime.Time.IsZero() {
		return &c.LastTransitionTime.Time
	}
	return nil
}

func getFinishTimeFromDeletionTimestamp(p *v1.Pod) *time.Time {
	if p.DeletionTimestamp != nil {
		finishTime := p.DeletionTimestamp.Time.Add(-time.Duration(pointer.Int64Deref(p.DeletionGracePeriodSeconds, 0)) * time.Second)
		return &finishTime
	}
	return nil
}

func (backoff backoffRecord) getRemainingTime(clock clock.WithTicker, defaultBackoff time.Duration, maxBackoff time.Duration) time.Duration {
	if backoff.failuresAfterLastSuccess == 0 {
		return 0
	}

	backoffDuration := defaultBackoff
	for i := 1; i < int(backoff.failuresAfterLastSuccess); i++ {
		backoffDuration = backoffDuration * 2
		if backoffDuration >= maxBackoff {
			backoffDuration = maxBackoff
			break
		}
	}

	timeElapsedSinceLastFailure := clock.Since(*backoff.lastFailureTime)

	if backoffDuration < timeElapsedSinceLastFailure {
		return 0
	}

	return backoffDuration - timeElapsedSinceLastFailure
}
