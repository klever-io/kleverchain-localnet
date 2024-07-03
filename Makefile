ifndef VALIDATORS_NUM
VALIDATORS_NUM := 3
endif

ifndef GCP_PROJECT
GCP_PROJECT := klever-blockchain-dev
endif

ifndef MAX_SUPPLY
MAX_SUPPLY := 10000000000000000
endif


create-localnet:
	VALIDATORS_NUM=${VALIDATORS_NUM} \
	MAX_SUPPLY=${MAX_SUPPLY} \
	CONSENSUS_GROUP_SIZE=${VALIDATORS_NUM} \
	python3 ./scripts/generate.py