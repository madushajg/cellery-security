package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"

	"cellery.io/cellery-security/components/envoy-oidc-filter/oidc"
	ext_authz "github.com/envoyproxy/go-control-plane/envoy/service/auth/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

const (
	ProviderUrlEnv             = "PROVIDER_URL"
	SkipDiscoveryCertVerifyEnv = "SKIP_DISCOVERY_URL_CERT_VERIFY"
	ClientIdEnv                = "CLIENT_ID"
	ClientSecretEnv            = "CLIENT_SECRET"
	RedirectUrlEnv             = "REDIRECT_URL"
	LogoutUrlEnv               = "LOGOUT_URL"
	AppUrlEnv                  = "APP_BASE_URL"
	DcrEpEnv                   = "DCR_ENDPOINT"
	DcrUser                    = "DCR_USER"
	DcrPassword                = "DCR_PASSWORD"
	NonSecurePaths             = "NON_SECURE_PATHS"
	SecurePaths                = "SECURE_PATHS"
	PrivateKeyFile             = "PRIVATE_KEY_FILE"
	CertificateFile            = "CERTIFICATE_FILE"
	JwtIssuer                  = "JWT_ISSUER"
	JwtAudience                = "JWT_AUDIENCE"
	SubjectClaim               = "SUBJECT_CLAIM"
	FilterListenerPort         = "FILTER_LISTENER_PORT"
	HttpCallbackListenerPort   = "HTTP_CALLBACK_LISTENER_PORT"
	CellNamespace              = "CELL_NAMESPACE"
)

func main() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt)

	_, skipCertVerify := os.LookupEnv(SkipDiscoveryCertVerifyEnv)
	if skipCertVerify {
		http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	}
	cfg := &oidc.Config{
		Provider:        os.Getenv(ProviderUrlEnv),
		ClientID:        os.Getenv(ClientIdEnv),
		ClientSecret:    os.Getenv(ClientSecretEnv),
		RedirectURL:     os.Getenv(RedirectUrlEnv),
		LogoutURL:       os.Getenv(LogoutUrlEnv),
		BaseURL:         os.Getenv(AppUrlEnv),
		DcrEP:           os.Getenv(DcrEpEnv),
		DcrUser:         os.Getenv(DcrUser),
		DcrPassword:     os.Getenv(DcrPassword),
		NonSecurePaths:  getNonSecurePaths(),
		SecurePaths:     getSecurePaths(),
		PrivateKeyFile:  os.Getenv(PrivateKeyFile),
		CertificateFile: os.Getenv(CertificateFile),
		JwtIssuer:       fmt.Sprintf("%s.%s", os.Getenv(JwtIssuer), os.Getenv(CellNamespace)),
		JwtAudience:     fmt.Sprintf("%s.%s", os.Getenv(JwtAudience), os.Getenv(CellNamespace)),
		SubjectClaim:    os.Getenv(SubjectClaim),
	}
	err := cfg.Validate()
	if err != nil {
		log.Fatal(err)
	}

	auth, err := oidc.NewAuthenticator(cfg)
	if err != nil {
		log.Fatal(err)
	}

	// Start auth callback http server
	go func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/_auth/callback", auth.Callback)
		mux.HandleFunc("/_auth/logout", auth.Logout)
		port := LookupEnv(HttpCallbackListenerPort, "15810")
		fmt.Printf("Starting HTTP auth callback reciver on %q\n", port)
		log.Fatal(http.ListenAndServe(":"+port, mux))
	}()

	// Start auth check envoy gRPC filter
	go func() {
		s := grpc.NewServer()
		ext_authz.RegisterAuthorizationServer(s, auth)
		reflection.Register(s)

		port := LookupEnv(FilterListenerPort, "15800")
		fmt.Printf("Starting gRPC filter reciver on %q\n", port)
		lis, err := net.Listen("tcp", ":"+port)
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}
		log.Fatal(s.Serve(lis))
	}()
	<-c
}

func LookupEnv(key string, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}

func getSecurePaths() []string {
	_, exist := os.LookupEnv(SecurePaths)
	if !exist || len(os.Getenv(SecurePaths)) == 0 {
		return nil
	}
	elems := strings.Split(os.Getenv(SecurePaths), ",")
	paths := make([]string, len(elems))
	for i, elem := range elems {
		paths[i] = strings.TrimSpace(elem)
	}
	fmt.Printf("secure paths: [ %v ] \n", paths)
	return paths
}

func getNonSecurePaths() []string {
	_, exist := os.LookupEnv(NonSecurePaths)
	if !exist || len(os.Getenv(NonSecurePaths)) == 0 {
		return nil
	}
	elems := strings.Split(os.Getenv(NonSecurePaths), ",")
	paths := make([]string, len(elems))
	for i, elem := range elems {
		paths[i] = strings.TrimSpace(elem)
	}
	fmt.Printf("Non secure paths: %v \n", paths)
	return paths
}
