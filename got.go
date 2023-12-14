package got

import (
	"context"
	"errors"
	"net/http"
	"time"
	"math/rand"
)

// Got holds got download config.
type Got struct {
	ProgressFunc

	Client *http.Client

	ctx context.Context
}


const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// ErrDownloadAborted - When download is aborted by the OS before it is completed, ErrDownloadAborted will be triggered
var ErrDownloadAborted = errors.New("Operation aborted")

// Seed for random userID
var seededRand *rand.Rand = rand.New(rand.NewSource(time.Now().UnixNano()))

func generateRandomString(length int) string {
    b := make([]byte, length)
    for i := range b {
        b[i] = charset[seededRand.Intn(len(charset))]
    }
    return string(b)
}

// DefaultClient is the default http client for got requests.
var DefaultClient = &http.Client{
	Transport: &http.Transport{
		MaxIdleConns:        10,
		IdleConnTimeout:     60 * time.Second,
		TLSHandshakeTimeout: 10 * time.Second,
		Proxy:               http.ProxyFromEnvironment,
	},
}

// Download creates *Download item and runs it.
func (g Got) Download(URL, dest string) error {

	return g.Do(&Download{
		ctx:    g.ctx,
		URL:    URL,
		Dest:   dest,
		Client: g.Client,
	})
}

// Do inits and runs ProgressFunc if set and starts the Download.
func (g Got) Do(dl *Download) error {

	if err := dl.Init(); err != nil {
		return err
	}

	if g.ProgressFunc != nil {

		defer func() {
			dl.StopProgress = true
		}()

		go dl.RunProgress(g.ProgressFunc)
	}

	return dl.Start()
}

// New returns new *Got with default context and client.
func New() *Got {
	return NewWithContext(context.Background())
}

// NewWithContext wants Context and returns *Got with default http client.
func NewWithContext(ctx context.Context) *Got {
	return &Got{
		ctx:    ctx,
		Client: DefaultClient,
	}
}

// NewRequest returns a new http.Request and error if any.
func NewRequest(ctx context.Context, method, URL string, header []GotHeader) (req *http.Request, err error) {

	if req, err = http.NewRequestWithContext(ctx, method, URL, nil); err != nil {
		return
	}

	userID := generateRandomString(16)

	req.Header.Set("User-Agent", userID)

	for _, h := range header {
		req.Header.Set(h.Key, h.Value)
	}

	return
}
