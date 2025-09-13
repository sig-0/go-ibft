.PHONY: lint
lint:
	golangci-lint run --config ./.github/golangci.yaml

.PHONY: gofumpt
gofumpt:
	go install mvdan.cc/gofumpt@latest
	gofumpt -l -w .

.PHONY: fixalign
fixalign:
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	fieldalignment -fix $(filter-out $@,$(MAKECMDGOALS)) # the full package name (not path!)

.PHONY: protoc
protoc:
	protoc --proto_path=./proto --go_out=./message --go_opt=module=github.com/sig-0/go-ibft/message ./proto/*.proto


