# Kubernetes WebHook Authentication/Authorization Minikube HelloWorld


Sample minimal HelloWorld application for Minikube demonstrating [Kubernetes Authentication/Authorization](https://kubernetes.io/docs/admin/kubelet-authentication-authorization/) as well as [Mutating and Validating AdmissionControl](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/) Webhooks.

Webhooks provide a mechanism for delegating k8s AU/AZ decisions.  In the case here, both policy decisions
are delegated to an _external_ HTTP REST service which you can test by running locally boomeranged through ngrok.  For more information on WebHooks:

- [WebHook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication)
- [WebHook Authorization](https://kubernetes.io/docs/admin/authorization/webhook/)

Mutating/Validating Webhooks allows you to modify or reject state changes through an external system

well, just read [What are admission webhooks?](https://kubernetes.io/docs/reference/access-authn-authz/extensible-admission-controllers/#what-are-admission-webhooks)


you have two users, `user1` and `user2` both of how have their own JWT HMAC credentials.

Once this app is deployed, everyone is authorized to access any api endpoint!!!....except user1 is not allowed to list pods....

i know, silly and arbitrary.

This repo demonstrates both using minikube and the external webhook deployed on GCP Cloud RUn

---

## Installation

### Deploy with Cloud Run


1. Download ngrok and run a default http proxy

```bash
git clone https://github.com/salrashid123/k8s_webhook_helloworld
cd k8s_webhook_helloworld/
cd server/
gcloud run deploy webhook-server --source . --region=us-central1 --allow-unauthenticated


Service [webhook-server] revision [webhook-server-00003-nz9] has been deployed and is serving 100 percent of traffic.
Service URL: https://webhook-server-995081019036.us-central1.run.app
```

So i've leved this app running at the url above

2. Set `authn.yaml`, `authz.yaml` to use ngrok

  Edit the two files and set the url appropriately

* `authn.yaml`

```yaml
apiVersion: v1
kind: Config
clusters:
  - name: my-authn-service
    cluster:
      server: https://webhook-server-995081019036.us-central1.run.app/authenticate
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
      server: https://webhook-server-995081019036.us-central1.run.app/authorize      
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

mkdir -p $HOME/.minikube/files/var/lib/minikube/certs/auth
cp authn.yaml $HOME/.minikube/files/var/lib/minikube/certs/auth
cp authz.yaml $HOME/.minikube/files/var/lib/minikube/certs/auth

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

Then invoke the API server with the k8s server endpoint IP and token using a user's JWT thats HMAC encoded (the hmackey is just "secret")

* `user1@domain.com`: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIxQGRvbWFpbi5jb20ifQ.W0Ek34LU4WQOxXdTqZ9Z-0kESz0wIEdYehxZHlTjt2I`

```json
{
  "alg": "HS256",
  "typ": "JWT"
}.
{
  "username": "user1@domain.com"
}
```

* `user2@domain.com`: `eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIyQGRvbWFpbi5jb20ifQ.DTvRw2dBVBOxOyt-Osq2e0iblh_xcbEy-Ir0ZBkkSdY`

```json
{
  "alg": "HS256",
  "typ": "JWT"
}.
{
  "username": "user2@domain.com"
}
```

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


* authenticate:

![images/authenticate.png](images/authenticate.png)


* deny:

![images/authorize_deny.png](images/authorize_deny.png)


* allow:

![images/authorize_allow.png](images/authorize_allow.png)

### Calling k8s with kubectl

If you want to use kubectl, you need to configure a context that will use either a  token or basic auth:

For example, the following `~/.kube/config` sets up two user contexts that you can use (`webhook1` and `webhook2`)

replace the minikube server ip with the value of `minikube ip`

```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/srashid/.minikube/ca.crt
    server: https://192.168.39.96:8443
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

### Mutating Webhook

First off, this code is nothing new...its basically just this: [simple-kubernetes-webhook](https://github.com/slackhq/simple-kubernetes-webhook).

Why am i rewriting it...sometimes i learn by taking something, modifying it and then commit it to a git library...this is all

if you want the canonical example, please see the link above...all i did here is copy

```bash
minikube stop
minikube delete
minikube start --driver=kvm2 

kubectl apply -f ns.yaml
kubectl apply -f webhook.yaml

kubectl get mutatingwebhookconfigurations,validatingwebhookconfigurations
```


The server output on apply is something like

```log
Starting Server..
received mutation request
INFO[0443] no lifespan label found, applying default lifespan toleration  min_lifespan=0 mutation=min_lifespan pod_name=offensive-pod
sending response
received validation request
INFO[0443] delete me called                              pod_name=offensive-pod
sending response
received mutation request
INFO[0497] no lifespan label found, applying default lifespan toleration  min_lifespan=0 mutation=min_lifespan pod_name=no-labels
sending response
received validation request
INFO[0497] delete me called                              pod_name=no-labels
sending response
received mutation request
INFO[0605] setting lifespan tolerations                  min_lifespan=3 mutation=min_lifespan pod_name=lifespan-three
sending response
received validation request
INFO[0605] delete me called                              pod_name=lifespan-three
sending response
received mutation request
INFO[0605] setting lifespan tolerations                  min_lifespan=7 mutation=min_lifespan pod_name=lifespan-seven
sending response
received validation request
INFO[0605] delete me called                              pod_name=lifespan-seven
sending response
```

so  the mutation and validation can be shown by looking at the pods

```bash
$ kubectl apply -f .
    pod/lifespan-seven created
    pod/lifespan-three created
    pod/no-labels created
    namespace/apps unchanged
    mutatingwebhookconfiguration.admissionregistration.k8s.io/simple-kubernetes-webhook.acme.com unchanged
    validatingwebhookconfiguration.admissionregistration.k8s.io/simple-kubernetes-webhook.acme.com unchanged
    Error from server: error when creating "bad-name.pod.yaml": admission webhook "simple-kubernetes-webhook.acme.com" denied the request: pod name contains "offensive"


$ kubectl get po -n apps 
NAME             READY   STATUS    RESTARTS   AGE
lifespan-seven   1/1     Running   0          6s
lifespan-three   1/1     Running   0          6s
no-labels        1/1     Running   0          115s
```

##### `Mutating`

The following will adjust tolerneces and inject an env-var to the pods

- `$ kubectl get po lifespan-three -n apps -oyaml`

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    acme.com/lifespan-requested: "3"          <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
  name: lifespan-three
  namespace: apps
spec:
  containers:
  - args:
    - sleep
    - "3600"
    env:
    - name: KUBE       <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
      value: "true"    <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
    image: busybox
    name: lifespan-three
  nodeName: minikube
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "14"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "13"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "12"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "11"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "10"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "9"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "8"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "7"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "6"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "5"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "4"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "3"
```

- `$ kubectl get po lifespan-seven -n apps -oyaml`

```yaml
apiVersion: v1
kind: Pod
metadata:
  labels:
    acme.com/lifespan-requested: "7"
  name: lifespan-seven
  namespace: apps
spec:
  containers:
  - args:
    - sleep
    - "3600"
    env:
    - name: KUBE       <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
      value: "true"    <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
    image: busybox
    name: lifespan-seven

  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "14"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "13"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "12"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "11"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "10"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "9"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "8"
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Equal
    value: "7"
```

- `$ kubectl get po no-labels -n apps -oyaml`

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: no-labels
  namespace: apps

spec:
  containers:
  - args:
    - sleep
    - "3600"
    env:
    - name: KUBE       <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
      value: "true"    <<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<<
    image: busybox
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  - effect: NoSchedule
    key: acme.com/lifespan-remaining
    operator: Exists
```


##### `Validating`

The following will prevent deployment of a pod if a 'bad name" is provided


```bash
$ kubectl apply -f bad-name.pod.yaml

  Error from server: error when creating "bad-name.pod.yaml": admission webhook "simple-kubernetes-webhook.acme.com" denied the request: pod name contains "offensive"
```

---

