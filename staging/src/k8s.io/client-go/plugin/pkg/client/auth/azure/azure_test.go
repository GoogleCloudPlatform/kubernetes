package azure

import (
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/Azure/go-autorest/autorest/azure"
)

func TestAzureTokenSource(t *testing.T) {
	fakeAccessToken := "fake token 1"
	fakeSource := fakeTokenSource{
		accessToken: fakeAccessToken,
		expiresOn:   strconv.FormatInt(time.Now().Add(3600*time.Second).Unix(), 10),
	}
	cfg := make(map[string]string)
	persiter := &fakePersister{cache: make(map[string]string)}
	tokenCache := newAzureTokenCache()
	tokenSource := newAzureTokenSource(&fakeSource, tokenCache, cfg, persiter)
	token, err := tokenSource.Token()
	if err != nil {
		t.Errorf("failed to retrieve the token form cache: %v", err)
	}

	wantCacheLen := 1
	if len(tokenCache.cache) != wantCacheLen {
		t.Errorf("Token() cache length error: got %v, want %v", len(tokenCache.cache), wantCacheLen)
	}

	if token != tokenCache.cache[azureTokenKey] {
		t.Error("Token() returned token != cached token")
	}

	wantCfg := token2Cfg(token)
	persistedCfg := persiter.Cache()
	for k, v := range persistedCfg {
		if strings.Compare(v, wantCfg[k]) != 0 {
			t.Errorf("Token() persisted cfg %s: got %v, want %v", k, v, wantCfg[k])
		}
	}

	fakeSource.accessToken = "fake token 2"
	token, err = tokenSource.Token()
	if err != nil {
		t.Errorf("failed to retrieve the cached token: %v", err)
	}

	if token.token.AccessToken != fakeAccessToken {
		t.Errorf("Token() didn't return the cached token")
	}
}

type fakePersister struct {
	lock  sync.Mutex
	cache map[string]string
}

func (p *fakePersister) Persist(cache map[string]string) error {
	p.lock.Lock()
	defer p.lock.Unlock()
	p.cache = map[string]string{}
	for k, v := range cache {
		p.cache[k] = v
	}
	return nil
}

func (p *fakePersister) Cache() map[string]string {
	ret := map[string]string{}
	p.lock.Lock()
	defer p.lock.Unlock()
	for k, v := range p.cache {
		ret[k] = v
	}
	return ret
}

type fakeTokenSource struct {
	expiresOn   string
	accessToken string
}

func (ts *fakeTokenSource) Token() (*azureToken, error) {
	return &azureToken{
		token:    newFackeAzureToken(ts.accessToken, ts.expiresOn),
		clientID: "fake",
		tenantID: "fake",
	}, nil
}

func token2Cfg(token *azureToken) map[string]string {
	cfg := make(map[string]string)
	cfg[cfgAccessToken] = token.token.AccessToken
	cfg[cfgRefreshToken] = token.token.RefreshToken
	cfg[cfgClientID] = token.clientID
	cfg[cfgTenantID] = token.tenantID
	cfg[cfgApiserverID] = token.token.Resource
	cfg[cfgExpiresIn] = token.token.ExpiresIn
	cfg[cfgExpiresOn] = token.token.ExpiresOn
	return cfg
}

func newFackeAzureToken(accessToken string, expiresOn string) azure.Token {
	return azure.Token{
		AccessToken:  accessToken,
		RefreshToken: "fake",
		ExpiresIn:    "3600",
		ExpiresOn:    expiresOn,
		NotBefore:    expiresOn,
		Resource:     "fake",
		Type:         "fake",
	}
}
