apt-get update
apt-get install -y golang-1.22 redis-tools


echo "# Set up environment variables" >> /etc/bash.bashrc
echo "export GCP_PROJECT=${projectid}" >> /etc/bash.bashrc
echo "export MEMORYSTORE_LOCATION=${region}" >> /etc/bash.bashrc
echo "export MEMORYSTORE_INSTANCE=${memorystore}" >> /etc/bash.bashrc
echo "export MEMORYSTORE_IP=${memorystore_ip}" >> /etc/bash.bashrc
echo "export MEMORYSTORE_PORT=${memorystore_port}" >> /etc/bash.bashrc
echo "export PATH=$PATH:/usr/lib/go-1.22/bin" >> /etc/bash.bashrc

echo "${memorystore_cert}" > /tmp/ca.crt
