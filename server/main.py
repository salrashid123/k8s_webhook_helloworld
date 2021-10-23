#!/usr/bin/python

from flask import Flask, request, Response, jsonify, abort
from functools import wraps

import logging, sys
import json

app = Flask(__name__)
app.config['DEBUG'] = True

import jwt

HMAC_SECRET='secret'

@app.route('/', methods = ['GET'])
def Default():
    resp = Response()
    resp.headers['Content-Type'] ='text/plain'
    return 'ok'

@app.route('/authenticate', methods = ['POST'])
def Authenticate():
    print('Requesting Authn' )
    parsed = json.loads(request.data)
    print(json.dumps(parsed, indent=4, sort_keys=True))
    token = parsed['spec']['token']
    
    resp = Response()
    resp.headers['Content-Type'] ='application/json'
    
    # determine the user givne bearer auth token and authenticate anyone
    try:
        decoded = jwt.decode(token, HMAC_SECRET, algorithms=['HS256'])
        username = decoded['username']
        r = {
                "apiVersion": "authentication.k8s.io/v1beta1", 
                "kind": "TokenReview", 
                "status": {
                    "authenticated": True, 
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
                        "username": username
                    }
                }
            }

    except Exception as de:
        r = {
                "apiVersion": "authentication.k8s.io/v1beta1", 
                "kind": "TokenReview", 
                "status": {
                    "authenticated": False
                }
            }
         
    return jsonify(r)

@app.route('/authorize', methods = ['POST'])
def Authorize():
    print('Requesting Authz' )
    parsed = json.loads(request.data)
    print(json.dumps(parsed, indent=4, sort_keys=True))
    
    # set some simple rules:
    # allow everyone except user1@domain.com when listing pods
    allowed = True
    user = parsed['spec']['user']
    
    try:
      resource = parsed['spec']['resourceAttributes']['resource']
      print(resource)
      if (user == 'user1@domain.com' and resource == 'pods'):
        allowed = False
    except KeyError:
      pass

    resp = Response()
    resp.headers['Content-Type'] ='application/json'
    r = {
            "apiVersion": "authorization.k8s.io/v1beta1", 
            "kind": "SubjectAccessReview", 
            "status": {
                "allowed": allowed
            }
        }

    return jsonify(r)

if __name__ == '__main__':
    #context = ('server.crt','server.key')
    #app.run(host='0.0.0.0', port=8081, debug=True,  threaded=True, ssl_context=context)
    app.run(host='0.0.0.0', port=8080, debug=True)
