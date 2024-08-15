.PHONY: clean hardhat bindings build rebuild

rebuild: clean hardhat bindings build

build: bin/safes

bin/safes:
	mkdir -p bin
	go build -o bin/safes ./

bindings/Safe/Safe.go:
	mkdir -p bindings/Safe
	seer evm generate --package Safe --output bindings/Safe/Safe.go --hardhat safe-smart-account/build/artifacts/contracts/Safe.sol/Safe.json --cli --struct Safe

bindings/SafeL2/SafeL2.go:
	mkdir -p bindings/SafeL2
	seer evm generate --package SafeL2 --output bindings/SafeL2/SafeL2.go --hardhat safe-smart-account/build/artifacts/contracts/SafeL2.sol/SafeL2.json --cli --struct SafeL2

bindings/SafeProxy/SafeProxy.go:
	mkdir -p bindings/SafeProxy
	seer evm generate --package SafeProxy --output bindings/SafeProxy/SafeProxy.go --hardhat safe-smart-account/build/artifacts/contracts/proxies/SafeProxy.sol/SafeProxy.json --cli --struct SafeProxy

bindings/SafeProxyFactory/SafeProxyFactory.go:
	mkdir -p bindings/SafeProxyFactory
	seer evm generate --package SafeProxyFactory --output bindings/SafeProxyFactory/SafeProxyFactory.go --hardhat safe-smart-account/build/artifacts/contracts/proxies/SafeProxyFactory.sol/SafeProxyFactory.json --cli --struct SafeProxyFactory

bindings: bindings/Safe/Safe.go bindings/SafeL2/SafeL2.go bindings/SafeProxy/SafeProxy.go bindings/SafeProxyFactory/SafeProxyFactory.go

clean:
	rm -rf bindings/* bin/*

hardhat:
	cd safe-smart-account && npm install && npx hardhat compile
