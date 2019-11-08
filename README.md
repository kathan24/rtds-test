# [Envoy](https://www.envoyproxy.io/) [RTDS](https://www.envoyproxy.io/docs/envoy/latest/api-v2/service/discovery/v2/rtds.proto#envoy-api-msg-service-discovery-v2-runtime) Prototype

This repo is forked from [Ivan Han's rtds-test](https://github.com/ivanhan/rtds-test). I have extended it to prototype
 - RTDS changes propagated to entire cluster rather than individual node.
 - Clear runtime value from Envoy sidecar once the purpose is fulfilled. 

## Prerequisites

- You will need the Envoy binary.

## Steps 
 - Clone this repo.
 - Launch three Envoy proxies (in different terminals), pointing it to the config found in this repo. E.g. assuming your Envoy binary, `envoy`, is located under `~/Downloads`, do:  

    ```sh
    ~/Downloads/envoy -c $GOPATH/src/github.com/kathan24/rtds-test/config/envoy_google.yaml --service-node node1 --service-cluster cluster1

    ~/Downloads/envoy -c $GOPATH/src/github.com/kathan24/rtds-test/config/envoy_facebook.yaml --service-node node1 --service-cluster cluster2

    ~/Downloads/envoy -c $GOPATH/src/github.com/kathan24/rtds-test/config/envoy_microsoft.yaml --service-node node2 --service-cluster cluster2
    ```

- This will run two clusters - `cluster1` and `cluster2`. 
    - `cluster1` has one node, node1
    - `cluster2` has two nodes, `node1` and `node2`

 - Verify that Envoy sidecar works as expected
    - `cluster1` with `node1`
        - `http://localhost:10000` - redirects-to / loads https://www.google.com
        - `http:localhost:9000` - you should see empty runtime values for Envoy
    - `cluster2` with `node1`
        - `http://localhost:10001` - redirects-to / loads https://www.facebook.com
        - `http:localhost:9001` - you should see empty runtime values for Envoy
    - `cluster2` with `node2`
        - `http://localhost:10002` - redirects-to / loads https://www.microsoft.com
        - `http:localhost:9002` - you should see empty runtime values for Envoy
- Launch RTDS management server
    ```sh
    go run $GOPATH/src/github.com/kathan24/rtds-test/main.go
    ```
- You should see
    - Within 10 seconds 
        - `cluster1` with `node1` should have NO effect 
        - `cluster2` with `node1` and `node2`
            - `http://localhost:10001` and `http://localhost:10002` - you should see `fault filter abort` with `HTTP status 404`
            - `http://localhost:9001` and `http://localhost:9002` - you should see runtime value as 
                ```
                {
                    "fault.http.abort.abort_percent": {
                        "layer_values": [
                            "100"
                        ],
                        "final_value": "100"
                    },
                    "fault.http.abort.http_status": {
                        "layer_values": [
                            "404"
                        ],
                        "final_value": "404"
                    }
                }
                ```
    - After 10 seconds 
        - Fault will be cleared and you should see the redirect happeneing again to `www.facebook.com` and `www.microsoft.com` for `node1` and `node2` of `cluster2` respectively. 