package sdknotifier

import (
	"context"
	"fmt"
	"net/http"

	"github.com/IsaacDSC/featureflag/pkg/ctxlog"
	"github.com/IsaacDSC/featureflag/pkg/pubsub"
)

type Subscriber interface {
	Listener(ctx context.Context, channel string, fn pubsub.Handler)
}

type SdkNotifyHandler struct {
	routes map[string]func(w http.ResponseWriter, r *http.Request)
	sub    Subscriber
}

func NewSdkNotifyHandler(sub Subscriber) *SdkNotifyHandler {
	h := new(SdkNotifyHandler)
	h.sub = sub
	h.routes = map[string]func(w http.ResponseWriter, r *http.Request){
		"GET /events/{resource}": h.event,
	}

	return h
}

func (h SdkNotifyHandler) GetRoutes() map[string]func(w http.ResponseWriter, r *http.Request) {
	return h.routes
}

func (h SdkNotifyHandler) event(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	ctx := r.Context()
	resource := r.PathValue("resource")
	if resource == "" {
		resource = "default"
	}

	l := ctxlog.GetLogger(ctx)
	l.Debug("connected sse", "resource", resource)

	h.sub.Listener(ctx, resource, func(ctx context.Context, msg pubsub.Msg) error {
		// var body Msg
		// if err := msg.ToJson(&body); err != nil {
		// 	return fmt.Errorf("error parser to json: %w", err)
		// }
		l.Info("SENT MSG")

		fmt.Fprintf(w, "data: %s\n\n", string(msg))
		w.(http.Flusher).Flush()

		return nil
	})

}
