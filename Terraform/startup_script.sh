apt-get update
apt-get install -y golang-1.22 redis-tools


echo "# Set up environment variables" >> /etc/bash.bashrc
echo "export GCP_PROJECT=${projectid}" >> /etc/bash.bashrc
echo "export MEMORYSTORE_LOCATION=${region}" >> /etc/bash.bashrc
echo "export MEMORYSTORE_INSTANCE=${memorystore}" >> /etc/bash.bashrc
echo "export PATH=$PATH:/usr/lib/go-1.22/bin" >> /etc/bash.bashrc
