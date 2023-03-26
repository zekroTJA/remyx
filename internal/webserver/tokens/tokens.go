package tokens

import (
	"encoding/json"
	"os"
	"time"

	"github.com/zekroTJA/timedmap"
	"github.com/zekrotja/remyx/internal/shared"
	"golang.org/x/oauth2"
)

const tokenCacheFile = ".tokencache"

type Cache interface {
	Set(key string, tk *oauth2.Token, exp time.Duration)
	Get(key string) (tk *oauth2.Token, ok bool)
	Close() error
}

type cache struct {
	m *timedmap.TimedMap
}

func NewCache() (Cache, error) {
	f, err := os.Open(tokenCacheFile)
	if err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	defer f.Close()

	m := make(map[string]*oauth2.Token)
	if f != nil {
		err = json.NewDecoder(f).Decode(&m)
		if err != nil {
			return nil, err
		}
	}

	tm, err := timedmap.FromMap(m, shared.SessionLifetime, 5*time.Minute)
	if err != nil {
		return nil, err
	}

	c := &cache{m: tm}
	return c, nil
}

func (t *cache) Set(key string, tk *oauth2.Token, exp time.Duration) {
	t.m.Set(key, tk, exp)
}

func (t *cache) Get(key string) (tk *oauth2.Token, ok bool) {
	tk, ok = t.m.GetValue(key).(*oauth2.Token)
	return tk, ok
}

func (t *cache) Close() error {
	f, err := os.Create(tokenCacheFile)
	if err != nil {
		return err
	}
	defer f.Close()

	snapshot := t.m.Snapshot()
	m := make(map[string]interface{})
	for k, v := range snapshot {
		ks, ok := k.(string)
		if !ok {
			continue
		}
		m[ks] = v
	}
	return json.NewEncoder(f).Encode(m)
}
