#!/bin/bash

#
# next set is only for setting up enterprise crunchy postgres repo
# not required if you build on centos
#
#sudo mkdir /opt/crunchy
#sudo cp $BUILDBASE/conf/crunchypg95.repo /etc/yum.repos.d
#sudo cp $BUILDBASE/conf/CRUNCHY* /opt/crunchy
#sudo yum -y install postgresql95-server

sudo yum -y install net-tools bind-utils wget unzip git 

#
# this set is required to build the docs
#
sudo yum -y install asciidoc ruby
gem install --pre asciidoctor-pdf
wget -O $HOME/bootstrap-4.5.0.zip http://laurent-laville.org/asciidoc/bootstrap/bootstrap-4.5.0.zip
asciidoc --backend install $HOME/bootstrap-4.5.0.zip
mkdir -p $HOME/.asciidoc/backends/bootstrap/js
#cp $GOPATH/src/github.com/crunchydata/crunchy-containers/docs/bootstrap.js \
#$HOME/.asciidoc/backends/bootstrap/js/
unzip $HOME/bootstrap-4.5.0.zip  $HOME/.asciidoc/backends/bootstrap/

