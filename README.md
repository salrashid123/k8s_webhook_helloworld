# Kubernetes WebHook Authentication/Authorization Minikube HelloWorld


Sample minimal HelloWorld application for Miikube demonstrating [Kubernetes Authentication/Authorization](https://kubernetes.io/docs/admin/kubelet-authentication-authorization/) using Webhooks.

Webhooks proivde a mechanism for delegating k8s AU/AZ decisions.  In the case here, both policy decisions
are delegated to an _external_ HTTP REST service.  For more information on WebHooks:

- [WebHook Authentication](https://kubernetes.io/docs/admin/authentication/#webhook-token-authentication)
- [WebHook Authorization](https://kubernetes.io/docs/admin/authorization/webhook/)

This repo is designed to run locally with minikube while the WebHook server may run locally or remotely as a separate
python Flask applicaton.  The steps detailed below are lengthy and involve copying certificates into your minikube' persistent
volume and running your webhook server.

You can deploy the AuthN/AuthZ server as a service within the k8s cluster and provide the Cluster DNS entry reference to it from the
webhook configuration files.  More information about that configuration in the appendix.

> __Note__: This is just a sample helloworld app.  Do not use this in production!

---

## Prerequsites

### Install python openssl
```
$ sudo apt-get update
$ sudo apt-get install python-pip openssl -y
$ sudo pip install virtualenv
```

## Installation

### Clone the repository

```
git clone salrashid123/k8s_webook_helloworld
```

### Prepare Minikube configuration for Webhook

Minikube needs to know the authn/authz config files and CA webhook_plugin certs to trust but those need to be accesible
while minikube is started.  That is, the config, certs and key needs to exist within the minikube VM while its starting up.
The easiest way to do this is to use [host mount folder](https://github.com/kubernetes/minikube/blob/master/docs/host_folder_mount.md) 
and reference the files on startup. However, not all minikube drivers support mount folders 
(as was the case for me; i used ```-vm-driver kvm2```.   One workaround employed for ```kvm2``` 
(and described in the appendix), is to first start minikube, then create the certificate files in a peristent volume
folder (eg [/var/lib/localkube/](https://kubernetes.io/docs/getting-started-guides/minikube/#persistent-volumes)


#### Start Minikube without custom configuration

```bash
$ minikube start

$ minikube ssh

$ sudo su -
```

### Find External interface IP address for your workstation

On your *workstation*, find its external interface's IP address

```
/sbin/ifconfig -a
wlan0: flags=4163<UP,BROADCAST,RUNNING,MULTICAST>  mtu 1500
        inet 10.17.134.45  netmask 255.255.224.0  broadcast 10.17.159.255
```

In the example above, the ip is ```10.17.134.45```

### Set host alias within minikube VM

While within the *minikube VM*   ```/etc/hosts``` file and add:
```
$ more /etc/hosts
127.0.0.1       localhost
127.0.1.1       minikube
10.17.134.45 webhook.domain.local
```

Minikube uses the host interface so this way, minkube can find the webhook server running on the workstation

> Note: you may have to reset the IP address within minikube's VM across restarts since the ```/etc/hosts``` value get overridden


#### Specify authn/authz definitions

Create the following files with the values shown below:

* /var/lib/localkube/authn.yaml
* /var/lib/localkube/authz.yaml


If you don't mind using the default hostname and certificates provided, run the following while in minikube:

```bash
curl -s  https://raw.githubusercontent.com/salrashid123/k8s_webook_helloworld/master/authn.yaml -o /var/lib/localkube/authn.yaml
curl -s  https://raw.githubusercontent.com/salrashid123/k8s_webook_helloworld/master/authz.yaml -o /var/lib/localkube/authz.yaml
curl -s  https://raw.githubusercontent.com/salrashid123/k8s_webook_helloworld/master/CA/CA_crt.pem -o /var/lib/localkube/certs/webhook_ca.crt
curl -s  https://raw.githubusercontent.com/salrashid123/k8s_webook_helloworld/master/CA/webhook_plugin.crt -o /var/lib/localkube/certs/webhook_plugin.crt
curl -s  https://raw.githubusercontent.com/salrashid123/k8s_webook_helloworld/master/CA/webhook_plugin.key -o /var/lib/localkube/certs/webhook_plugin.key
```

and skip to the "Stop Minikube" section below, otherwise:

- authn.yaml 

```yaml
clusters:
  - name: my-authn-service
    cluster:
      certificate-authority: /var/lib/localkube/certs/webhook_ca.crt
      server: https://webhook.domain.local:8081/authenticate

users:
  - name: my-api-server
    user:
      client-certificate: /var/lib/localkube/certs/webhook_plugin.crt
      client-key: /var/lib/localkube/certs/webhook_plugin.key

current-context: webhook
contexts:
- context:
    cluster: my-authn-service
    user: my-api-sever
  name: webhook
```

- authz.yaml 
```yaml
clusters:
  - name: my-authz-service
    cluster:
      certificate-authority: /var/lib/localkube/certs/webhook_ca.crt
      server: https://webhook.domain.local:8081/authorize

users:
  - name: my-api-server
    user:
      client-certificate: /var/lib/localkube/certs/webhook_plugin.crt
      client-key: /var/lib/localkube/certs/webhook_plugin.key

current-context: webhook
contexts:
- context:
    cluster: my-authz-service
    user: my-api-sever
  name: webhook
```

#### Specify CA trust certificate and plugin certs

>> __NOTE__: specifying the CA and certficates here properly is critical!

Create the following files:

* /var/lib/localkube/certs/webhook_ca.crt
  * default: [webhook_ca.crt](CA/CA_crt.pem)
* /var/lib/localkube/certs/webhook_plugin.crt
  * default: [webhook_plugin.crt](CA/webhook_plugin.crt)
* /var/lib/localkube/certs/webhook_plugin.key
  * default: [webhook_plugin.key](CA/webhook_plugin.key)

You can use the default files provided in this repo

#### Stop Minikube

```
minikube stop
```

Since the config and cert files are saved under ```/var/lib/localkube/```, they will remain after minikube restarts


## Start Webhook Server and Minikube


### Start WebHook Server

Now run the webhook server:

```
cd server
virtualenv env
source env/bin/activate
pip install -r requirements
```

Copy the certificates generated
```
cp ../CA/server.crt .
cp ../CA/server.key .
```

Start the server:
```
python webhook.py
```

The default certificates for the webhook server in this git repo uses

- CN: ```CN=webhook.domain.local``` 

- SAN: ```X509v3 Subject Alternative Name:  DNS: webhook.domain.local```

You can change the CN and SAN specifications as well as define your own custom CA.  Instructions for that can be found in the Appendix.


### Restart Minikube with webhook

Now Restart Minikube with the webhook configurations

```bash
$ minikube start --extra-config apiserver.Authentication.WebHook.ConfigFile=/var/lib/localkube/authn.yaml \
    --extra-config apiserver.Authorization.Mode=Webhook \
    --extra-config apiserver.Authorization.WebhookConfigFile=/var/lib/localkube/authz.yaml
```

You can verify connectivity from minikube --> webhook server by entering minikube and testing the endpoint:

Make sure you can connect from the VM to the webhook server by name:
```
minikube ssh

curl -vk https://webhook.domain.com:8081/
```

If that does not succeed, check if ```/etc/hosts``` file within the VM is set to the external IP address of your workstation

### Verify Webhook callback

Minikube is now configured to talk to the local webhook server.  Verification can be done
by either directly invoking the k8s API server or via kubectl.

### Calling k8s with Bearer Token directly

First step is we need to know the endpoint of the minikube-k8s sever:

to do that simply run

```bash
$ minikube status
minikube: Running
cluster: Running
kubectl: Correctly Configured: pointing to minikube-vm at 192.168.39.196
```

Then invoke the API server with the k8s server endpoint IP and token:

```bash
curl -vk \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VybmFtZSI6InVzZXIxQGRvbWFpbi5jb20ifQ.W0Ek34LU4WQOxXdTqZ9Z-0kESz0wIEdYehxZHlTjt2I" \
  https://192.168.39.196:8443/api/
```

All endpoints are authorized for the token above.  The only endpoint that isn't is ```/api/pods```

The token above can be anything becuase the sample Webhook approves all requests for ```/authenticate``` and ```/authorize```

### Calling k8s with kubectl

If you want to use kubectl, you need to configure a context that will use either a  token or basic auth:

For example, the following ```~/.kube/config``` sets up two user contexts that you can use (```webhook1``` and ```webhookd2```)
```yaml
apiVersion: v1
clusters:
- cluster:
    certificate-authority: /home/srashid/.minikube/ca.crt
    server: https://192.168.39.196:8443
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

Verify you can still access cluter informaton
```bash
$ kubectl get no
NAME       STATUS    ROLES     AGE       VERSION
minikube   Ready     <none>    2d        v1.9.0
```

### Verify 

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

#### Authenticaiton Response

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

#### Authoriztion Request
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
        "nonResourceAttributes": {
            "path": "/api/", 
            "verb": "get"
        }, 
        "uid": "42", 
        "user": "user1@yourdomain.com"
    }, 
    "status": {
        "allowed": false
    }
}
```

#### Authoriztion Response
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

The following section is optional and details how you can override the certificates to create
your own CA and populate the server's certificate with a CN and SAN of your choosing:

### Create webhook server certificates

- Create your own CA:

```bash
 openssl genrsa -out CA_key.pem 2048
 
 openssl req -x509 -new -nodes -key CA_key.pem -out CA_crt.pem -config openssl.cnf \
   -subj "/C=US/ST=California/L=Mountain View/O=Google/OU=Enterprise/CN=MyCA"
```

- Edit openssl.cnf and add in the DNS/IP you want to use 

```bash
cd k8s_webook_helloworld/certs
vi openssl.cnf
```

set the [SAN values](https://github.com/salrashid123/k8s_webook_helloworld/blob/master/CA/openssl.cnf#L114):
```
[alt_names]
#IP.1 = 35.202.144.185
DNS.1 = webhook.domain.local
DNS.2 = webhook-srv
DNS.3 = webhook-srv.kube-system
DNS.4 = webhook-srv.kube-system.svc
DNS.5 = webhook-srv.kube-system.svc.cluster.local
```

> Note: DNS2->5 are there to support having webhook server hosted as a k8s service (```webhook-svc```)

- Create the server certificate and specify the CN= inline
```bash
openssl genrsa -out server.key 2048

openssl req -config openssl.cnf -out server.csr -key server.key -new -sha256 \
   -subj "/C=US/ST=California/L=Mountain View/O=Google/OU=Enterprise/CN=webhook.domain.local"

openssl ca -config openssl.cnf -days 365 -notext  -in server.csr   -out server.crt
```
Check the certificate created has the corect CN and SAN values:
```bash
openssl x509 -in server.crt  -text -noout
```

see:
```
    CN = webhook.domain.local

        X509v3 extensions:
            Netscape Comment: 
                OpenSSL Generated Certificate
            X509v3 Subject Alternative Name: 
                DNS:webhook.domain.local, DNS:webhook-srv, DNS:webhook-srv.kube-system, DNS:webhook-srv.kube-system.svc, DNS:webhook-srv.kube-system.svc.cluster.local
            X509v3 Key Usage: 
                Digital Signature, Non Repudiation, Key Encipherment
```


- Generate the plugin certificate

```bash
openssl genrsa -out webhook_plugin.key 2048
openssl req -config openssl.cnf -out webhook_plugin.csr -key webhook_plugin.key -new -sha256 \
   -subj "/C=US/ST=California/L=Mountain View/O=Google/OU=Enterprise/CN=webhook_plugin.default.cluster.local"
```

```bash
openssl ca -config openssl.cnf -days 365 -notext  -in webhook_plugin.csr   -out webhook_plugin.crt
```

### IP address for SAN

You can also use an IP address as the endpoint if you do not have an external DNS server available.

To use an IP instead, change the SAN value to include it and regenerate the server.crt/key

edit CA/openssl.cnf and add IP.1 value
```
[alt_names]
IP.1  = 35.202.144.185
DNS.1 = webhook.domain.local
DNS.2 = webhook-srv
DNS.3 = webhook-srv.kube-system
DNS.4 = webhook-srv.kube-system.svc
DNS.5 = webhook-srv.kube-system.svc.cluster.local
```

You should see:

```
X509v3 Subject Alternative Name: 
    IP Address:35.202.144.185, DNS:webhook-srv, DNS:webhook-srv.kube-system, DNS:webhook-srv.kube-system.svc, DNS:webhook-srv.kube-system.svc.cluster.local
```
### Creating new users

The sample here used HMAC JWT to create and decode the tokens:

```python
import jwt
encoded = jwt.encode({'username': 'user1@domain.com'}, 'secret', algorithm='HS256')
print encoded

decoded = jwt.decode(encoded, 'secret', algorithms=['HS256'])
print decoded
```

## Troubleshooting

- Verify connectivity from minikube VM to your AU/AZ server:
```
minikube ssh
```

then from within minikube, check connectivity:

```
$ curl -vk https://webhook.domain.com:8081/
```

If that fails, then there maybe several reasons:

* Local firewall/iptables
  If you have iptabbles running, it may interfere with vm->host routing. Please disable iptable rules if AU/AZ is running
  on the local workstation

* Check ```/etc/hosts``` value within the VM and set it to the external IP address of your workstaiton
* Provider```--vm-driver kvm2```


## References

- [https://github.com/pasientskyhosting/kubernetes-authserver](https://github.com/pasientskyhosting/kubernetes-authserver)
