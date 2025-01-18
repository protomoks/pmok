package recorder

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/protomoks/pmok/internal/config"
	"github.com/protomoks/pmok/internal/mockspec"
	"github.com/protomoks/pmok/internal/mockspec/json"
	"github.com/protomoks/pmok/internal/mockspec/mimetypes"
	"github.com/protomoks/pmok/internal/utils/constants"
)

type RecordCommand struct {
	Target        string
	ResponsesPath string
}

func (c RecordCommand) Valid() error {
	if c.Target == "" {
		return errors.New("target is required")
	}
	return nil
}

func Run(ctx context.Context, command RecordCommand) error {
	if err := command.Valid(); err != nil {
		return err
	}

	if err := config.CreateMocksDirIfNotExist(command.ResponsesPath); err != nil {
		return err
	}

	rec := &recorder{
		targetUrl:  command.Target,
		targetDir:  filepath.Join(config.MocksDir, command.ResponsesPath),
		ctx:        ctx,
		workerChan: make(chan targetResponse),
	}

	server := http.Server{
		Addr:    fmt.Sprintf(":%d", constants.RecorderDefaultPort),
		Handler: rec,
	}

	// start the background worker
	done := make(chan bool)
	go rec.processResponses(done)

	serverErr := make(chan error, 1)
	// start the recorder
	go func() {
		serverErr <- server.ListenAndServe()
	}()

	// wait...
	select {
	case <-ctx.Done():
		fmt.Println("Shutting down server")
		ctxShutdown, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := server.Shutdown(ctxShutdown); err != nil {
			return fmt.Errorf("failed to shut down recorder gracefully %w", err)
		}
		// stop the worker
		close(done)
	case err := <-serverErr:
		close(done)
		return fmt.Errorf("server error %w", err)
	}
	return nil
}

type recorder struct {
	targetUrl  string
	targetDir  string
	ctx        context.Context
	workerChan chan targetResponse
}

type targetResponse struct {
	response *http.Response
}

func (rec *recorder) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Received a request")
	url := rec.targetUrl + r.URL.Path
	proxyr, err := http.NewRequestWithContext(rec.ctx, r.Method, url, r.Body)
	if err != nil {
		fmt.Printf("Error when creating request to %s\n", url)
		return
	}
	client := &http.Client{}
	proxyr.Header = r.Header

	res, err := client.Do(proxyr)
	if err != nil {
		fmt.Printf("Error when receiving response for %s . Error %w\n", url, err)
		return
	}

	fmt.Printf("Header %s\n", res.Header.Get("Content-Type"))
	clonedRes, err := cloneResponse(res)
	if err != nil {
		fmt.Printf("Error when cloning response for %s\n", url)
		return
	}

	defer res.Body.Close()

	rec.workerChan <- targetResponse{
		response: clonedRes,
	}
	io.Copy(w, res.Body)

}

func cloneResponse(res *http.Response) (*http.Response, error) {
	var bodyBuf bytes.Buffer
	tee := io.TeeReader(res.Body, &bodyBuf)
	// populate the bodyBuf with the body response
	_, err := io.ReadAll(tee)
	if err != nil {
		return nil, err
	}
	res.Body.Close()
	// restore the body
	res.Body = io.NopCloser(&bodyBuf)
	// clone the response
	clone := *res
	// add response to the clone
	clone.Body = io.NopCloser(bytes.NewBuffer(bodyBuf.Bytes()))
	return &clone, nil
}

func (rec *recorder) processResponses(done <-chan bool) {
	for {
		select {
		case res := <-rec.workerChan:
			mw, err := rec.getMockWriterFromTargetResponse(res.response)
			if err != nil {
				continue
			}
			if err := mw.WriteResponse(res.response); err != nil {
				continue
			}
			mw.Close()
		case <-done:
			fmt.Println("worker shutting down")
			return
		}
	}
}

func (rec *recorder) getMockWriterFromTargetResponse(res *http.Response) (mockspec.MockWriter, error) {
	contentType := res.Header.Get("Content-Type")
	var mw mockspec.MockWriter
	name := filepath.Join(rec.targetDir, mockspec.MockFileNameFromPath(res.Request.URL.Path))

	switch contentType {
	case mimetypes.ContentTypeJSON:
		file, err := os.Create(name)
		if err != nil {
			return nil, err
		}
		mw = json.New(file)
	}

	return mw, nil
}
