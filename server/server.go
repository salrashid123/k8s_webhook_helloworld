package main

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/golang-jwt/jwt"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/http2"
	authenticationv1 "k8s.io/api/authentication/v1"

	"github.com/salrashid123/k82_webhook_helloworld/server/webhook/admission"
	admissionv1 "k8s.io/api/admission/v1"
	"k8s.io/kubernetes/pkg/apis/authorization"
)

var (
	port = flag.String("port", ":8080", "Port to listen on; default :8080")
)

const (
	hmacSecret = "secret"
)

func authenticate(w http.ResponseWriter, r *http.Request) {

	var treq *authenticationv1.TokenReview
	status := &authenticationv1.TokenReviewStatus{}

	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&treq)
	if err != nil {
		fmt.Printf("error parsing TokenReview  %v\n", err)
	} else {

		prettyJSON, err := json.MarshalIndent(treq, "", "  ")
		if err != nil {
			fmt.Printf("JSON parse error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("Authn Request %s\n", string(prettyJSON))

		token, err := jwt.Parse(treq.Spec.Token, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("Unexpected signing method: %v", token.Header["alg"])
			}
			return []byte(hmacSecret), nil
		})
		if err != nil {
			fmt.Printf("Error validating auth token %v", err)
		} else {
			if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
				status = &authenticationv1.TokenReviewStatus{
					Authenticated: true,
					User: authenticationv1.UserInfo{
						Username: claims["username"].(string),
						UID:      "42",
						Extra:    map[string]authenticationv1.ExtraValue{"extrafield1": {"extravalue1", "extravalue2"}},
						Groups:   []string{"developers,qa"},
					},
				}
			} else {
				fmt.Printf("error validating TokenRequest signature  %v\n", err)
			}
		}
	}

	resp := struct {
		APIVersion string                              `json:"apiVersion"`
		Kind       string                              `json:"kind"`
		Status     *authenticationv1.TokenReviewStatus `json:"status"`
	}{
		APIVersion: treq.APIVersion, // authenticationv1.SchemeGroupVersion.String(),
		Kind:       "TokenReview",
		Status:     status,
	}

	prettyJSON, err := json.MarshalIndent(treq, "", "  ")
	if err != nil {
		fmt.Printf("JSON parse error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Authn Response %s\n", string(prettyJSON))

	json.NewEncoder(w).Encode(resp)
}

func authorize(w http.ResponseWriter, r *http.Request) {

	var req authorization.SubjectAccessReview
	allowed := false
	w.Header().Set("Content-Type", "application/json")

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		fmt.Printf("error parsing SubjectAccessReview  %v\n", err)
	} else {

		prettyJSON, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			fmt.Printf("JSON parse error: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fmt.Printf("Authz Request %s\n", string(prettyJSON))

		if req.Spec.ResourceAttributes != nil {
			if req.Spec.User == "user1@domain.com" && req.Spec.ResourceAttributes.Resource == "pods" {
				fmt.Println("denying user1@domain.com for pods")
				allowed = false
			} else {
				allowed = true
			}
		}
	}
	type status struct {
		Allowed         bool   `json:"allowed"`
		Reason          string `json:"reason"`
		EvaluationError string `json:"evaluationError"`
	}
	resp := struct {
		APIVersion string `json:"apiVersion"`
		Kind       string `json:"kind"`
		Status     status `json:"status"`
	}{
		APIVersion: req.APIVersion, //"authorization.k8s.io/v1beta1", //authorizationv1.SchemeGroupVersion.String(),
		Kind:       "SubjectAccessReview",
		Status:     status{Allowed: allowed},
	}

	prettyJSON, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		fmt.Printf("JSON parse error: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	fmt.Printf("Authz Response %s\n", string(prettyJSON))

	json.NewEncoder(w).Encode(resp)
}

func gethandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, fmt.Sprintf("ok"))
}

func validate(w http.ResponseWriter, r *http.Request) {

	fmt.Println("received validation request")

	in, err := parseRequest(*r)
	if err != nil {
		fmt.Printf("could not generate admission response: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	logger := logrus.WithField("uri", r.RequestURI)

	logger.Debug("received validation request")

	adm := admission.Admitter{
		Logger:  logger,
		Request: in.Request,
	}

	out, err := adm.ValidatePodReview()
	if err != nil {
		fmt.Printf("could not generate admission response: %v", err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jout, err := json.Marshal(out)
	if err != nil {
		fmt.Printf("could not parse admission response: %v", err)
		http.Error(w, fmt.Sprintf("could not parse admission response: %v", err), http.StatusInternalServerError)
		return
	}

	fmt.Println("sending response")
	fmt.Fprintf(w, "%s", jout)
}

func mutate(w http.ResponseWriter, r *http.Request) {
	fmt.Println("received mutation request")

	in, err := parseRequest(*r)
	if err != nil {
		fmt.Printf("could not parse admission response: %v", err)
		http.Error(w, fmt.Sprintf("could not parse admission response: %v", err), http.StatusInternalServerError)
		return
	}
	logger := logrus.WithField("uri", r.RequestURI)
	logger.Debug("received validation request")
	adm := admission.Admitter{
		Logger:  logger,
		Request: in.Request,
	}

	out, err := adm.MutatePodReview()
	if err != nil {
		e := fmt.Sprintf("could not generate admission response: %v", err)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jout, err := json.Marshal(out)
	if err != nil {
		e := fmt.Sprintf("could not parse admission response: %v", err)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	fmt.Println("sending response")
	fmt.Fprintf(w, "%s", jout)
}

func parseRequest(r http.Request) (*admissionv1.AdmissionReview, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("Content-Type: %q should be %q",
			r.Header.Get("Content-Type"), "application/json")
	}

	bodybuf := new(bytes.Buffer)
	bodybuf.ReadFrom(r.Body)
	body := bodybuf.Bytes()

	if len(body) == 0 {
		return nil, fmt.Errorf("admission request body is empty")
	}

	var a admissionv1.AdmissionReview

	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("could not parse admission review request: %v", err)
	}

	if a.Request == nil {
		return nil, fmt.Errorf("admission review can't be used: Request field is nil")
	}

	return &a, nil
}

func main() {

	flag.Parse()

	router := mux.NewRouter()
	router.Methods(http.MethodGet).Path("/").HandlerFunc(gethandler)
	router.Methods(http.MethodPost).Path("/authenticate").HandlerFunc(authenticate)
	router.Methods(http.MethodPost).Path("/authorize").HandlerFunc(authorize)
	router.Methods(http.MethodPost).Path("/mutate").HandlerFunc(mutate)
	router.Methods(http.MethodPost).Path("/validate").HandlerFunc(validate)
	clientCaCert, err := ioutil.ReadFile("tls-ca-chain.pem")
	clientCaCertPool := x509.NewCertPool()
	clientCaCertPool.AppendCertsFromPEM(clientCaCert)

	config := &tls.Config{
		// ClientAuth: tls.RequireAndVerifyClientCert,
		// ClientCAs: clientCaCertPool,
	}

	var server *http.Server
	server = &http.Server{
		Addr:      *port,
		Handler:   router,
		TLSConfig: config,
	}
	http2.ConfigureServer(server, &http2.Server{})
	fmt.Println("Starting Server..")
	//err = server.ListenAndServeTLS("server.crt", "server.key")
	err = server.ListenAndServe()
	fmt.Printf("Unable to start Server %v", err)

}
