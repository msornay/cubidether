#!/bin/bash 

set -e

sudo add-apt-repository -y ppa:ethereum/ethereum
sudo apt-get update
sudo apt-get install cpp-ethereum fglrx-updates python-twisted

sudo mkdir /opt/ethproxy && chown $USER eth-proxy && cd eth-proxy 
git clone https://github.com/Atrides/eth-proxy .
cat <<EOF >> eth-proxy.conf
    COIN = "ETH"
    HOST = "0.0.0.0"
    PORT = 8080
    WALLET = "{{.Coinbase}}"
    ENABLE_WORKER_ID = False
    MONITORING = False
    POOL_HOST = "eth-eu.dwarfpool.com"
    POOL_PORT = "8008"
    POOL_FAILOVER_ENABLE = True
    POOL_HOST_FAILOVER1 = "us1.ethermine.org"
    POOL_PORT_FAILOVER1 = 4444
    POOL_HOST_FAILOVER2 = "us2.ethermine.org"
    POOL_PORT_FAILOVER2 = 4444
    POOL_HOST_FAILOVER3 = "asia1.ethermine.org"
    POOL_PORT_FAILOVER3 = 4444
    LOG_TO_FILE = True
    DEBUG = False
EOF

echo "The Ether must flow !"

