package service

import (
	"encoding/json"
	"fmt"
	"time"

	"drill-platform/internal/domain/entity"
)

type RedisClient interface {
	Get(key string) (string, error)
	Set(key string, value interface{}, expiration time.Duration) error
	Delete(keys ...string) error
}

const stepCacheTTL = 2 * time.Hour
const userCacheTTL = 10 * time.Minute

func stepCacheKey(drillID uint64) string {
	return fmt.Sprintf("drill:steps:%d", drillID)
}

const userNamesCacheKey = "drill:user_names"

func GetCachedSteps(redis RedisClient, drillID uint64) ([]entity.StepInstance, bool) {
	if redis == nil {
		return nil, false
	}
	data, err := redis.Get(stepCacheKey(drillID))
	if err != nil {
		return nil, false
	}
	var steps []entity.StepInstance
	if err := json.Unmarshal([]byte(data), &steps); err != nil {
		return nil, false
	}
	return steps, true
}

func SetCachedSteps(redis RedisClient, drillID uint64, steps []entity.StepInstance) {
	if redis == nil {
		return
	}
	data, err := json.Marshal(steps)
	if err != nil {
		return
	}
	redis.Set(stepCacheKey(drillID), data, stepCacheTTL)
}

func InvalidateStepCache(redis RedisClient, drillID uint64) {
	if redis == nil {
		return
	}
	redis.Delete(stepCacheKey(drillID))
}

func PatchCachedStep(redis RedisClient, drillID uint64, stepID uint, patch map[string]interface{}) {
	if redis == nil {
		return
	}
	steps, ok := GetCachedSteps(redis, drillID)
	if !ok {
		return
	}
	for i, s := range steps {
		if s.ID == uint64(stepID) {
			if v, ok := patch["status"].(string); ok {
				steps[i].Status = v
			}
			if v, ok := patch["start_time"].(*string); ok && v != nil {
				t, err := time.Parse(time.RFC3339, *v)
				if err == nil {
					steps[i].StartTime = &t
				}
			}
			if v, ok := patch["end_time"].(*string); ok && v != nil {
				t, err := time.Parse(time.RFC3339, *v)
				if err == nil {
					steps[i].EndTime = &t
				}
			}
			if v, ok := patch["timeout_at"].(*string); ok && v != nil {
				t, err := time.Parse(time.RFC3339, *v)
				if err == nil {
					steps[i].TimeoutAt = &t
				}
			}
			if v, ok := patch["remark"].(string); ok {
				steps[i].Remark = v
			}
			if v, ok := patch["issue_desc"].(string); ok {
				steps[i].IssueDesc = v
			}
			if v, ok := patch["assignee_names"].(string); ok {
				steps[i].AssigneeNames = v
			}
			break
		}
	}
	SetCachedSteps(redis, drillID, steps)
}

func GetCachedUserNames(redis RedisClient, ids []uint64) map[uint64]string {
	if redis == nil {
		return nil
	}
	data, err := redis.Get(userNamesCacheKey)
	if err != nil {
		return nil
	}
	var allNames map[uint64]string
	if err := json.Unmarshal([]byte(data), &allNames); err != nil {
		return nil
	}
	result := make(map[uint64]string, len(ids))
	for _, id := range ids {
		if name, ok := allNames[id]; ok {
			result[id] = name
		}
	}
	if len(result) == len(ids) {
		return result
	}
	return nil
}

func SetCachedUserNames(redis RedisClient, users []entity.User) {
	if redis == nil {
		return
	}
	data, err := redis.Get(userNamesCacheKey)
	var allNames map[uint64]string
	if err == nil {
		json.Unmarshal([]byte(data), &allNames)
	}
	if allNames == nil {
		allNames = make(map[uint64]string, len(users))
	}
	for _, u := range users {
		allNames[u.ID] = u.RealName
	}
	newData, _ := json.Marshal(allNames)
	redis.Set(userNamesCacheKey, newData, userCacheTTL)
}
