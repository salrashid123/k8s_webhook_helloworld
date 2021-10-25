# Kubernetes WebHook Authentication/Authorization Minikube HelloWorld


Sample minimal HelloWorld application for Minikube demonstrating [Kubernetes Authentication/Authorization](https://kubernetes.io/docs/admin/kubelet-authentication-authorization/) using Webhooks.

Webhooks provide a mechanism for delegating k8s AU/AZ decisions.  In the case here, both policy decisions
are delegated to an _external_ HTTP REST service which you can test by running locally boomeranged through ngrok.  For more information on WebHooks:

- [WebHook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication)
- [WebHook Authorization](https://kubernetes.io/docs/admin/authorization/webhook/)

Wait, why is ngrok involved here?  I thought you said it was all local with minikube?

Yeah, i know, i just didn't  know how to make minikube call a url on the host system directly (the host meaning the laptop).

so this is where this becomes lazy and crap:  i make minikube/k8s call an "external" api server with a public ngrok url.  That url is basically a tunnel back to the laptop...

think of it as a boomerang.   If any reader can tell me how to make minikube talk to the local host/system/laptop, let me know

> __Note__: This is just a sample helloworld app; just a demo

besides, its also very old...there are much better examples around..i'd suggest following something lik

* [Kubernetes authentication/authorization webhook using golang in minikube](https://github.com/dinumathai/auth-webhook-sample#)
- [https://github.com/pasientskyhosting/kubernetes-authserver](https://github.com/pasientskyhosting/kubernetes-authserver)


The example is beyond silly.

you have two users, `user1` and `user2` both of how have their own JWT HMAC credentials.

Once this app is deployed, everyone is authorized to access any api endpoint!!!....except user1 is not allowed to list pods....

i know, silly and arbitrary.

---

## Installation

You can either `A)` deploy the external authn/authz server locally with `ngrok`

### Clone the repository

```bash
git clone https://github.com/salrashid123/k8s_webhook_helloworld
```


### Deploy with ngrok

You can test this locally to if you use a external proxy like [ngrok](https://ngrok.com/).

1. Download ngrok and run a default http proxy

```bash
./ngrok http -host-header=rewrite  localhost:8080
```
  
  This will assign a random url for you to use for 2hours (the free edition)

  In my case it was `https://2723-72-83-67-174.ngrok.io`:

  ![images/ngrok_url.png](images/ngrok_url.png)

  you can view the traffic by going to [http://localhost:4040/inspect/http](http://localhost:4040/inspect/http)

2. Set `authn.yaml`, `authz.yaml` to use ngrok

  Edit the two files and set the url appropriately

* `authn.yaml`

```yaml
apiVersion: v1
kind: Config
clusters:
  - name: my-authn-service
    cluster:
      server: https://2723-72-83-67-174.ngrok.io/authenticate
users:
  - name: my-api-server
    user:
      token: test-token
current-context: webhook
contexts:
- context:
    cluster: my-authn-service
    user: my-api-sever
  name: webhook
```

* `authz.yaml`

```yaml
apiVersion: v1
kind: Config
clusters:
  - name: my-authz-service
    cluster:
      server: https://2723-72-83-67-174.ngrok.io/authorize      
users:
  - name: my-api-server
    user:
      token: test-token
current-context: webhook
contexts:
- context:
    cluster: my-authz-service
    user: my-api-sever
  name: webhook
```

#### Specify authn/authz definitions

Edit the following files and enter in your `PROJECT_ID`.  

These kubernetes config files signals where to look for the external authn and authz configurations.

It also specifies the certificates to use when contacting GCP

- authn.yaml 

```yaml
apiVersion: v1
kind: Config
clusters:
  - name: my-authn-service
    cluster:
      certificate-authority: /var/lib/minikube/certs/gcp_roots.pem
      server: https://webhook-dot-mineral-minutia-820.appspot.com/authenticate
users:
  - name: my-api-server
    user:
      token: test-token
current-context: webhook
contexts:
- context:
    cluster: my-authn-service
    user: my-api-sever
  name: webhook
```

- authz.yaml 
```yaml
apiVersion: v1
kind: Config
clusters:
  - name: my-authz-service
    cluster:
      certificate-authority: /var/lib/minikube/certs/gcp_roots.pem
      server: https://webhook-dot-mineral-minutia-820.appspot.com/authorize
users:
  - name: my-api-server
    user:
      token: test-token
current-context: webhook
contexts:
- context:
    cluster: my-authz-service
    user: my-api-sever
  name: webhook
```

### Start Minikube with custom configuration

```bash
minikube stop
minikube delete

wget -O gcp_roots.pem https://pki.google.com/roots.pem

mkdir -p $HOME/.minikube/files/var/lib/minikube/certs/auth
cp authn.yaml $HOME/.minikube/files/var/lib/minikube/certs/auth
cp authz.yaml $HOME/.minikube/files/var/lib/minikube/certs/auth

cp gcp_roots.pem $HOME/.minikube/files/var/lib/minikube/certs/gcp_roots.pem
cp certs/webhook_ca.crt $HOME/.minikube/files/var/lib/minikube/certs/webhook_ca.crt

minikube start --driver=kvm2 --embed-certs \
   --extra-config apiserver.authorization-mode=RBAC,Webhook \
   --extra-config apiserver.authentication-token-webhook-config-file=/var/lib/minikube/certs/auth/authn.yaml \
   --extra-config apiserver.authorization-webhook-config-file=/var/lib/minikube/certs/auth/authz.yaml
```

(i used `--driver=kvm2`, you can certainly use whatever you want)

### Verify Webhook callback

Minikube is now configured to talk to the local webhook server.  Verification can be done
by either directly invoking the k8s API server or via kubectl.

### Calling k8s with Bearer Token directly

First step is we need to know the endpoint of the minikube-k8s sever:

to do that simply run

```bash
$ minikube ip
```

Then invoke the API server with the k8s server endpoint IP and token using a user's JWT thats HMAC encoded

* `user1@domain.com`: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIxQGRvbWFpbi5jb20ifQ.W0Ek34LU4WQOxXdTqZ9Z-0kESz0wIEdYehxZHlTjt2I`
* `user2@domain.com`: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIyQGRvbWFpbi5jb20ifQ.DTvRw2dBVBOxOyt-Osq2e0iblh_xcbEy-Ir0ZBkkSdY`

The authorization server is silly: it allows any user to anything *except* user1 to access pods...yeah, it silly but its my example


as user1 to see pods

```bash
curl -sk \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIxQGRvbWFpbi5jb20ifQ.W0Ek34LU4WQOxXdTqZ9Z-0kESz0wIEdYehxZHlTjt2I" \
  https://`minikube ip`:8443/api/v1/namespaces/default/pods

{
  "kind": "Status",
  "apiVersion": "v1",
  "metadata": {
    
  },
  "status": "Failure",
  "message": "pods is forbidden: User \"user1@domain.com\" cannot list resource \"pods\" in API group \"\" in the namespace \"default\"",
  "reason": "Forbidden",
  "details": {
    "kind": "pods"
  },
  "code": 403
}
```

The above bearer JWT is in the form:

```
{
  "alg": "HS256",
  "typ": "JWT"
}.
{
  "username": "user1@domain.com"
}
```

As user2 to see pods
```bash
curl -sk \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIyQGRvbWFpbi5jb20ifQ.DTvRw2dBVBOxOyt-Osq2e0iblh_xcbEy-Ir0ZBkkSdY" \
  https://`minikube ip`:8443/api/v1/namespaces/default/pods

{
  "kind": "PodList",
  "apiVersion": "v1",
  "metadata": {
    "resourceVersion": "1161"
  },
  "items": []
}
```

If you are using the `ngrok` console, you will see 

* authenticate:

![images/authenticate.png](images/authenticate.png)


* deny:

![images/authorize_deny.png](images/authorize_deny.png)


* allow:

![images/authorize_allow.png](images/authorize_allow.png)
### Calling k8s with kubectl

If you want to use kubectl, you need to configure a context that will use either a  token or basic auth:

For example, the following ```~/.kube/config``` sets up two user contexts that you can use (```webhook1``` and ```webhookd2```)

replace the minikube server ip with the value of `minikube ip`

```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/srashid/.minikube/ca.crt
    server: https://192.168.39.37:8443
  name: minikube
contexts:
- context:
    cluster: minikube
    user: minikube
  name: minikube
  
- context:
    cluster: minikube
    user: user1
  name: webhook1

- context:
    cluster: minikube
    user: user2
  name: webhook2

current-context: webhook1
kind: Config
preferences: {}
users:
- name: minikube
  user:
    as-user-extra: {}
    client-certificate: /home/srashid/.minikube/client.crt
    client-key: /home/srashid/.minikube/client.key
- name: user1
  user:
    token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIxQGRvbWFpbi5jb20ifQ.W0Ek34LU4WQOxXdTqZ9Z-0kESz0wIEdYehxZHlTjt2I
- name: user2
  user:
    token: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIyQGRvbWFpbi5jb20ifQ.DTvRw2dBVBOxOyt-Osq2e0iblh_xcbEy-Ir0ZBkkSdY
```

Swap contexts:
```bash
$ kubectl config use-context  webhook1
Switched to context "webhook1".
```

Verify you can still access cluster info partially as user1

```bash
$ kubectl get no
NAME       STATUS   ROLES                  AGE   VERSION
minikube   Ready    control-plane,master   18m   v1.22.2

$ kubectl get po
Error from server (Forbidden): pods is forbidden: User "user1@domain.com" cannot list resource "pods" in API group "" in the namespace "default"
```

Verify you can still access cluster info fully as user2

Either way, you should see the Authentication and Authorization requests in the window where your webhook server is running:

#### Authentication Request

```json
{
    "apiVersion": "authentication.k8s.io/v1beta1", 
    "kind": "TokenReview", 
    "metadata": {
        "creationTimestamp": null
    }, 
    "spec": {
        "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIxQGRvbWFpbi5jb20ifQ.W0Ek34LU4WQOxXdTqZ9Z-0kESz0wIEdYehxZHlTjt2I"
    }, 
    "status": {
        "user": {}
    }
}
```

#### Authentication Response

```json
{
    "apiVersion": "authentication.k8s.io/v1beta1", 
    "kind": "TokenReview", 
    "status": {
        "authenticated": true, 
        "user": {
            "extra": {
                "extrafield1": [
                    "extravalue1", 
                    "extravalue2"
                ]
            }, 
            "groups": [
                "developers", 
                "qa"
            ], 
            "uid": "42", 
            "username": "user1@yourdomain.com"
        }
    }
}
```

#### Authorization Request
```json
{
    "apiVersion": "authorization.k8s.io/v1beta1", 
    "kind": "SubjectAccessReview", 
    "metadata": {
        "creationTimestamp": null
    }, 
    "spec": {
        "extra": {
            "extrafield1": [
                "extravalue1", 
                "extravalue2"
            ]
        }, 
        "group": [
            "developers", 
            "qa", 
            "system:authenticated"
        ], 
        "resourceAttributes": {
            "namespace": "default",
            "resource": "pods",
            "verb": "list",
            "version": "v1"
        }, 
        "uid": "42", 
        "user": "user1@yourdomain.com"
    }, 
    "status": {
        "allowed": false
    }
}
```

#### Authorization Response
```json
{
    "apiVersion": "authorization.k8s.io/v1beta1", 
    "kind": "SubjectAccessReview", 
    "status": {
        "allowed": true
    }
}
```
## Appendix

### Creating new users

The sample here used HMAC JWT to create and decode the tokens:

```python
import jwt
encoded = jwt.encode({'username': 'user1@domain.com'}, 'secret', algorithm='HS256')
print encoded

decoded = jwt.decode(encoded, 'secret', algorithms=['HS256'])
print decoded
```

### Testing using mTLS

This is much harder and i've left this out and i've yet to get it to work

The authn and authz configurations _should_ allow mtls from the api server to the external webhook server (i think)

Unfortunately, i haven't gotten it to work yet..you're welcome to try, i think it involves enabling following flags

* mTLS:

```yaml
      client-certificate: /var/lib/minikube/certs/webhook_plugin.crt
      client-key: /var/lib/minikube/certs/webhook_plugin.key
```

When you configure minikube, copy these certs in

```bash
cp certs/webhook_plugin.crt $HOME/.minikube/files/var/lib/minikube/certs/webhook_plugin.crt
cp certs/webhook_plugin.key $HOME/.minikube/files/var/lib/minikube/certs/webhook_plugin.key
```

and in `main.py`.  Note the server certificate i specified there for TLS is `webhook.esodemoapp2.com` which resolves to an IP I (it wont work for you unless you do a lot of tricks with DNS and minikube)

```python
if __name__ == '__main__':
    context = ssl.SSLContext(ssl.PROTOCOL_TLSv1_2)
    context.verify_mode = ssl.CERT_REQUIRED
    context.verify_flags
    context.load_verify_locations('tls-ca-chain.pem')
    context.load_cert_chain('server.crt', 'server.key')
    app.run(host='0.0.0.0', port=8081, debug=True,  threaded=True, ssl_context=context)
```

```bash
curl -v -H "Host: webhook.esodemoapp2.com" \
  --cacert webhook_ca.crt \
  --cert webhook_plugin.crt \
  --key webhook_plugin.key \
   https://webhook.esodemoapp2.com:8081/
```