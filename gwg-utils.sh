#!/usr/bin/bash

set -e


SERVER_DIR       = "/etc/wireguard/"
WG_MANAGER_DIR   = $SERVER_DIR + ".wg_manager"
USERS_CONFIG_DIR = $SERVER_DIR + ".configs"
USERS_DIR        = $SERVER_DIR + "users"

command=$1

function preinstallGwg {
    echo "Installing Wireguard Server..."
    sudo apt install -y wireguard

    echo "Preparing system..."
    sudo groupadd gwg-manager
    sudo usermod -aG gwg-manager $USER
    sudo chown root:gwg-manager /etc/wireguard
    sudo chmod ug+rwx /etc/wireguard

    echo "Set gwg PATH..."
    sudo sh -c "echo export PATH=$PATH:/usr/bin/gwg >> /etc/profile"
    source /etc/profile

    echo "Enable ipv4 forwarding..."
    sudo sh -c "echo net.ipv4.ip_forward=1 >> /etc/sysctl.conf"
    sudo sysctl -p

    echo "Set gwg..."
    sudo mv gwg /usr/bin
    echo "Done"

    gwg version

    su - u $USER ./gwg-utils server_install
}

function postinstallGwg {
    echo "Creating gwg directory..."
    mkdir $WG_MANAGER_DIR
    mkdir $USERS_CONFIG_DIR
    mkdir $USERS_DIR

    echo "Creating template files..."
    mv wg_template.conf $WG_MANAGER_DIR && mv client_template.conf $WG_MANAGER_DIR

    echo "Installing wg server..."
    gwg install

    read -p 'You must log out to complete the installation. Ready [Y] ?' user
    echo
    sudo pkill -9 -u $USER
}

function updateGwg {
    wget https://github.com/PavelMilanov/go-wg-manager/releases/download/latest/gwg.tar
    tar -xzf gwg.tar
    sudo mv gwg /usr/bin
    rm gwg.tar
    mv wg_template.conf $WG_MANAGER_DIR && mv client_template.conf $WG_MANAGER_DIR
}

case "$command" in
    install)
        preinstallGwg;;
    server_install)
        postinstallGwg;;
    update)
        updateGwg;;
esac