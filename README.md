# Kubernetes WebHook Authentication/Authorization Minikube HelloWorld


Sample minimal HelloWorld application for Minikube demonstrating [Kubernetes Authentication/Authorization](https://kubernetes.io/docs/admin/kubelet-authentication-authorization/) using Webhooks.

Webhooks provide a mechanism for delegating k8s AU/AZ decisions.  In the case here, both policy decisions
are delegated to an _external_ HTTP REST service which i happened to run on Appengine for simplicity.  For more information on WebHooks:

- [WebHook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication)
- [WebHook Authorization](https://kubernetes.io/docs/admin/authorization/webhook/)

You can deploy the AuthN/AuthZ server as a service within the k8s cluster and provide the Cluster DNS entry reference to it from the
webhook configuration files.  More information about that configuration in the appendix.

> __Note__: This is just a sample helloworld app.  Do not use this in production!


...its also very old...there are much better examples around..i'd suggest following something lik

* [Kubernetes authentication/authorization webhook using golang in minikube](https://github.com/dinumathai/auth-webhook-sample#)
- [https://github.com/pasientskyhosting/kubernetes-authserver](https://github.com/pasientskyhosting/kubernetes-authserver)


The example is beyond silly.

you have two users, `user1` and `user2` both of how have their own JWT HMAC credentials.

Once this app is deployed, everyone is authorized to access any api endpoint!!!....except user1 is not allowed to list pods....

i know, silly and arbitrary.

---

## Installation

### Clone the repository

```
git clone https://github.com/salrashid123/k8s_webhook_helloworld
```

### Deploy WebHook Server

It is **MUCH** easier to deploy your webhook server to AppEngine than to run locally.  The latter requires you to play some
games with the ```/etc/hosts``` file for name resolution and to match the SAN values within the server certificate.  Running 
the webhook server locally is described in the Appendix.


#### Deploy to Google Appengine

Set up a google cloud platform project and enable AppeEngine

```bash
cd k8s_webhook_helloworld/server

pip install -t lib -r requirements.txt
```

now deploy the webhook server:

```
gcloud app deploy --version 1 --no-promote
```

At this point, your Webhook server should be accessible at:

```
gcloud config  get-value core/project

curl https://1-dot-webhook-dot-YOUR_PROJECT.appspot.com/
```

### Specify authn/authz definitions

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
      # token: test-token
      client-certificate: /var/lib/minikube/certs/webhook_plugin.crt
      client-key: /var/lib/minikube/certs/webhook_plugin.key
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
      # token: test-token
      client-certificate: /var/lib/minikube/certs/webhook_plugin.crt
      client-key: /var/lib/minikube/certs/webhook_plugin.key
current-context: webhook
contexts:
- context:
    cluster: my-authz-service
    user: my-api-sever
  name: webhook
```

### Prepare Minikube configuration for Webhook

Minikube needs to know the authn/authz config files and CA webhook_plugin certs to trust but those need to be accessible
while minikube is started.  That is, the config, certs and key needs to exist within the minikube VM while its starting up.
The easiest way to do this is to copy the files over to minikube's mapped directory `$HOME/.minikube/files/var/lib/minikube/certs/`.


#### Start Minikube with custom configuration

https://github.com/dinumathai/auth-webhook-sample#deploy-in-minikube


```bash
minikube stop
minikube delete

wget -O gcp_roots.pem https://pki.google.com/roots.pem

mkdir -p $HOME/.minikube/files/var/lib/minikube/certs/auth
cp authn.yaml $HOME/.minikube/files/var/lib/minikube/certs/auth
cp authz.yaml $HOME/.minikube/files/var/lib/minikube/certs/auth

cp gcp_roots.pem $HOME/.minikube/files/var/lib/minikube/certs/gcp_roots.pem
cp webhook_ca.crt $HOME/.minikube/files/var/lib/minikube/certs/webhook_ca.crt
cp webhook_plugin.crt $HOME/.minikube/files/var/lib/minikube/certs/webhook_plugin.crt
cp webhook_plugin.key $HOME/.minikube/files/var/lib/minikube/certs/webhook_plugin.key

minikube start --driver=kvm2 --embed-certs \
   --extra-config apiserver.authorization-mode=RBAC,Webhook \
   --extra-config apiserver.authentication-token-webhook-config-file=/var/lib/minikube/certs/auth/authn.yaml \
   --extra-config apiserver.authorization-webhook-config-file=/var/lib/minikube/certs/auth/authz.yaml
```

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

![alt text](images/gae_auth_logs.png)

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
