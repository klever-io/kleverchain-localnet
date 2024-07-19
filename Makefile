ifndef VALIDATORS_NUM
VALIDATORS_NUM := 1
endif

ifndef MAX_SUPPLY
MAX_SUPPLY := 10000000000000000
endif

generate_keys:
	docker run --rm -v $(shell pwd)/keys:/opt/klever-blockchain \
    	--user "$(shell id -u):$(shell id -g)" \
    	--entrypoint='' kleverapp/klever-go:latest keygenerator --num-keys ${VALIDATORS_NUM} --key-type both

create-localnet:
	VALIDATORS_NUM=${VALIDATORS_NUM} \
	MAX_SUPPLY=${MAX_SUPPLY} \
	CONSENSUS_GROUP_SIZE=${VALIDATORS_NUM} \
	python3 ./scripts/generate.py

compose-up:
	docker-compose up -d