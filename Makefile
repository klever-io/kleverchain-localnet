ifndef VALIDATORS_NUM
VALIDATORS_NUM := 1
endif

ifndef MAX_SUPPLY
MAX_SUPPLY := 10000000000000000
endif

generate_keys:
	VALIDATORS_NUM=${VALIDATORS_NUM} \
	MAX_SUPPLY=${MAX_SUPPLY} \
	CONSENSUS_GROUP_SIZE=${VALIDATORS_NUM} \
	python3 ./scripts/generate_keys.py

create-localnet:
	VALIDATORS_NUM=${VALIDATORS_NUM} \
	MAX_SUPPLY=${MAX_SUPPLY} \
	CONSENSUS_GROUP_SIZE=${VALIDATORS_NUM} \
	python3 ./scripts/generate.py

compose-up:
	docker-compose up -d