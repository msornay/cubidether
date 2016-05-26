#!/bin/bash 

set -e

INSTALL_DIR="$HOME/.ethproxy"

sudo apt-get update
#sudo apt-get install -y -q software-properties-common
#sudo add-apt-repository -y ppa:ethereum/ethereum
#sudo add-apt-repository -y ppa:ethereum/ethereum-qt
#sudo apt-get update
sudo apt-get install -y -q cpp-ethereum fglrx-updates python-twisted git supervisor

if [ ! -d $INSTALL_DIR ]
then
	mkdir -p $INSTALL_DIR
fi

cd $INSTALL_DIR

if [ -d eth-proxy/.git ] 
then
	echo "Already a git repo there, pull !"
	cd eth-proxy
	git pull
	cd ..
else
	git clone https://github.com/Atrides/eth-proxy
fi

cat <<EOF > eth-proxy/eth-proxy.conf
COIN = "ETH"
HOST = "0.0.0.0"
PORT = 8080
WALLET = "{{.Coinbase}}"
ENABLE_WORKER_ID = False
MONITORING = False
POOL_HOST = "eu1.ethermine.org"
POOL_PORT = 4444
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

cat <<EOF > /etc/supervisor/conf.d/eth-proxy.conf 
[program:eth-proxy]
command=/usr/bin/python $INSTALL_DIR/eth-proxy/eth-proxy.py
directory=$INSTALL_DIR
user=$USER
autostart=true
stderr_logfile=$INSTALL_DIR/eth-proxy.log
stdout_logfile=$INSTALL_DIR/eth-proxy.log
EOF

sudo supervisorctl reread && sudo supervisorctl start eth-proxy

echo "The Ether must flow !"

