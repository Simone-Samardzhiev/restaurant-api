package rate

import (
	"menu/internal/config"
	"strconv"

	"github.com/valkey-io/valkey-go"
	"golang.org/x/net/context"
)

var script = valkey.NewLuaScript(
	`local current = redis.call("INCR", KEYS[1])
if current == 1 then
    redis.call("EXPIRE", KEYS[1], ARGV[1])
end

if current > tonumber(ARGV[2]) then
    return -1
else
    return current
end`,
)

type ValkeyStore struct {
	client valkey.Client
	limit  string
	window string
}

func NewValkeyStore(client valkey.Client, c config.RateLimit) *ValkeyStore {
	return &ValkeyStore{
		client: client,
		limit:  strconv.Itoa(c.Limit),
		window: strconv.Itoa(int(c.Window.Seconds())),
	}
}

func (s *ValkeyStore) Allow(identifier string) (bool, error) {
	id := "rate_limit:" + identifier

	res := script.Exec(context.Background(), s.client, []string{id}, []string{s.window, s.limit})
	val, err := res.ToInt64()
	if err != nil {
		return false, err
	}

	return val != -1, nil
}
