package presence

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrUnavailable = errors.New("presence store unavailable")

// Redis supplies one clock for all API nodes; a single-key script also works in Cluster.
const activityScript = `
local clock = redis.call('TIME')
local now = tonumber(clock[1]) * 1000 + math.floor(tonumber(clock[2]) / 1000)
redis.call('ZREMRANGEBYSCORE', KEYS[1], '-inf', now - 300000)
if ARGV[1] ~= '' then
  local previous = tonumber(redis.call('ZSCORE', KEYS[1], ARGV[1])) or 0
  redis.call('ZADD', KEYS[1], math.max(now, previous), ARGV[1])
  redis.call('PEXPIRE', KEYS[1], 300000)
  return 1
end
return redis.call('ZRANGE', KEYS[1], 0, -1)
`

type Store struct {
	client redis.UniversalClient
	key    string
}

func NewStore(client redis.UniversalClient) *Store {
	return &Store{client: client, key: "drill:presence:users"}
}

func (s *Store) Touch(ctx context.Context, userID uint64) error {
	if s == nil || s.client == nil {
		return ErrUnavailable
	}
	if userID == 0 {
		return errors.New("presence requires an authenticated user")
	}
	ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	return s.client.Eval(ctx, activityScript, []string{s.key}, strconv.FormatUint(userID, 10)).Err()
}

func (s *Store) OnlineIDs(ctx context.Context) ([]uint64, error) {
	if s == nil || s.client == nil {
		return nil, ErrUnavailable
	}
	ctx, cancel := context.WithTimeout(ctx, 250*time.Millisecond)
	defer cancel()
	members, err := s.client.Eval(ctx, activityScript, []string{s.key}, "").StringSlice()
	if err != nil {
		return nil, err
	}
	ids := make([]uint64, 0, len(members))
	for _, member := range members {
		id, err := strconv.ParseUint(member, 10, 64)
		if err != nil || id == 0 {
			return nil, errors.New("invalid presence member")
		}
		ids = append(ids, id)
	}
	return ids, nil
}
