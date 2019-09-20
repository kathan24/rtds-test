# rtds-test

## Steps

1. Clone this repo.
2. Launch your Envoy proxy, pointing it to the config found in this repo. E.g. assuming your Envoy proxy executable, `envoy`, is located under `~/Downloads`, do:  

```
~/Downloads/envoy -c $GOPATH/src/github.com/ivanhan/rtds-test/config/envoy.yaml --service-node node1 --service-cluster cluster1
```

3. Navigate to http://localhost:10000 and see that it correctly navigates to https://www.google.com.
4. Launch the RTDS management server:

```
go run $GOPATH/src/github.com/ivanhan/rtds-test/main.go
```
5. Navigate to http://localhost:1000 and refresh the page several times. Notice that 50% of the time, you will get an abort error.
6. Navigate to http://localhost:1001/runtime to see the fault injection settings.
