package handlers
import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

)

type checkGroup struct {
	build string
	log   *zap.SugaredLogger
	db    *sqlx.DB
}


func (cg checkGroup) readiness(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), time.Second)
	defer cancel()

	status := "ok"
	statusCode := http.StatusOK
	if err := database.StatusCheck(ctx, cg.db); err != nil {
		status = "db not ready"
		statusCode = http.StatusInternalServerError
	}

	data := struct {
		Status string `json:"status"`
	}{
		Status: status,
	}

	if err := response(w, statusCode, data); err != nil {
		cg.log.Errorw("readiness", "ERROR", err)
	}

	cg.log.Infow("readiness", "statusCode", statusCode, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)
}


func (cg checkGroup) liveness(w http.ResponseWriter, r *http.Request) {
	host, err := os.Hostname()
	if err != nil {
		host = "unavailable"
	}

	data := struct {
		Status    string `json:"status,omitempty"`
		Build     string `json:"build,omitempty"`
		Host      string `json:"host,omitempty"`
		Pod       string `json:"pod,omitempty"`
		PodIP     string `json:"podIP,omitempty"`
		Node      string `json:"node,omitempty"`
		Namespace string `json:"namespace,omitempty"`
	}{
		Status:    "up",
		Build:     cg.build,
		Host:      host,
		Pod:       os.Getenv("KUBERNETES_PODNAME"),
		PodIP:     os.Getenv("KUBERNETES_NAMESPACE_POD_IP"),
		Node:      os.Getenv("KUBERNETES_NODENAME"),
		Namespace: os.Getenv("KUBERNETES_NAMESPACE"),
	}

	statusCode := http.StatusOK
	if err := response(w, statusCode, data); err != nil {
		cg.log.Errorw("liveness", "ERROR", err)
	}

	cg.log.Infow("liveness", "statusCode", statusCode, "method", r.Method, "path", r.URL.Path, "remoteaddr", r.RemoteAddr)
}

func response(w http.ResponseWriter, statusCode int, data interface{}) error {

	// Convert the response value to JSON.
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}


	w.Header().Set("Content-Type", "application/json")

	w.WriteHeader(statusCode)


	if _, err := w.Write(jsonData); err != nil {
		return err
	}

	return nil
}

