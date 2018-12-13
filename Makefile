BIN_DIR=_output/bin
RELEASE_VER=0.1
HOME=/home/xieyd

clean:
	rm -rf _output/
	rm -rf vendor/

vender-init:
	go run ${GOPATH}/src/github.com/kardianos/govendor/main.go init
	go run ${GOPATH}/src/github.com/kardianos/govendor/main.go add +e

vender-update:
	go run ${GOPATH}/src/github.com/kardianos/govendor/main.go update +e

crd-test:
	go run ${GOPATH}/src/github.com/xieydd/kubenetes-crd/test/kube-crd.go --kubeconfig=${HOME}/.kube/config