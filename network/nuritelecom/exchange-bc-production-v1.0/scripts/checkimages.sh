function displayLogo() {
         
    echo " ███╗   ██╗██╗   ██╗██████╗ ██╗    ██████╗ ██╗      ██████╗  ██████╗██╗  ██╗ ██████╗██╗  ██╗ █████╗ ██╗███╗   ██╗"
    echo " ████╗  ██║██║   ██║██╔══██╗██║    ██╔══██╗██║     ██╔═══██╗██╔════╝██║ ██╔╝██╔════╝██║  ██║██╔══██╗██║████╗  ██║"
    echo " ██╔██╗ ██║██║   ██║██████╔╝██║    ██████╔╝██║     ██║   ██║██║     █████╔╝ ██║     ███████║███████║██║██╔██╗ ██║"
    echo " ██║╚██╗██║██║   ██║██╔══██╗██║    ██╔══██╗██║     ██║   ██║██║     ██╔═██╗ ██║     ██╔══██║██╔══██║██║██║╚██╗██║"
    echo " ██║ ╚████║╚██████╔╝██║  ██║██║    ██████╔╝███████╗╚██████╔╝╚██████╗██║  ██╗╚██████╗██║  ██║██║  ██║██║██║ ╚████║"
    echo " ╚═╝  ╚═══╝ ╚═════╝ ╚═╝  ╚═╝╚═╝    ╚═════╝ ╚══════╝ ╚═════╝  ╚═════╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝╚═╝  ╚═╝╚═╝╚═╝  ╚═══╝"
}

function checkImages() {
  echo "hyperledger/fabric image check start..."

  if [[ "$(docker images -q hyperledger/fabric-ca:1.4.3 2> /dev/null)" == "" ]]; then
    echo "hyperledger/fabric-ca - invalid image version"
    exit 1
  fi

  if [[ "$(docker images -q hyperledger/fabric-zookeeper:0.4.15 2> /dev/null)" == "" ]]; then
    echo "hyperledger/fabric-zookeeper - invalid image version"
    exit 1
  fi

  if [[ "$(docker images -q hyperledger/fabric-kafka:0.4.15 2> /dev/null)" == "" ]]; then
    echo "hyperledger/fabric-kafka - invalid image version"
    exit 1
  fi

  if [[ "$(docker images -q hyperledger/fabric-orderer:1.4.3 2> /dev/null)" == "" ]]; then
    echo "hyperledger/fabric-orderer - invalid image version"
    exit 1
  fi

  if [[ "$(docker images -q hyperledger/fabric-peer:1.4.3 2> /dev/null)" == "" ]]; then
    echo "hyperledger/fabric-peer - invalid image version"
    exit 1
  fi

  if [[ "$(docker images -q hyperledger/fabric-tools:1.4.3 2> /dev/null)" == "" ]]; then
    echo "hyperledger/fabric-tools - invalid image version"
    exit 1
  fi

  echo "hyperledger/fabric image check completed successfully!!!"
}

function clearAllContainers() {
  docker stop $(docker ps -aq)
  docker rm $(docker ps -aq)
  echo "All containers cleared!!!"
}

function clearContainers() {
  CONTAINER_IDS=$(docker ps -a | awk '($2 ~ /dev-peer.*.exchange.*/) {print $1}')
  if [ -z "$CONTAINER_IDS" -o "$CONTAINER_IDS" == " " ]; then
    echo "---- No containers available for deletion ----"
  else
    docker rm -f $CONTAINER_IDS
  fi
}

function removeUnwantedImages() {
  DOCKER_IMAGE_IDS=$(docker images | awk '($1 ~ /dev-peer.*.exchange.*/) {print $3}')
  if [ -z "$DOCKER_IMAGE_IDS" -o "$DOCKER_IMAGE_IDS" == " " ]; then
    echo "---- No images available for deletion ----"
  else
    docker rmi -f $DOCKER_IMAGE_IDS
    echo "Removed unwanted images"
  fi
}

function checkPrereqs() {
  # Note, we check configtxlator externally because it does not require a config file, and peer in the
  # docker image because of FAB-8551 that makes configtxlator return 'development version' in docker
  echo "Check Prereqs"

  IMAGETAG="latest"
  LOCAL_VERSION=$(../bin/configtxlator version | sed -ne 's/ Version: //p')
  DOCKER_IMAGE_VERSION=$(docker run --rm hyperledger/fabric-tools:$IMAGETAG peer version | sed -ne 's/ Version: //p' | head -1)

  echo "LOCAL_VERSION=$LOCAL_VERSION"
  echo "DOCKER_IMAGE_VERSION=$DOCKER_IMAGE_VERSION"

  if [ "$LOCAL_VERSION" != "$DOCKER_IMAGE_VERSION" ]; then
    echo "=================== WARNING ==================="
    echo "  Local fabric binaries and docker images are  "
    echo "  out of  sync. This may cause problems.       "
    echo "==============================================="
  fi

  for UNSUPPORTED_VERSION in $BLACKLISTED_VERSIONS; do
    echo "$LOCAL_VERSION" | grep -q $UNSUPPORTED_VERSION
    if [ $? -eq 0 ]; then
      echo "ERROR! Local Fabric binary version of $LOCAL_VERSION does not match this newer version of BYFN and is unsupported. Either move to a later version of Fabric or checkout an earlier version of fabric-samples."
      exit 1
    fi

    echo "$DOCKER_IMAGE_VERSION" | grep -q $UNSUPPORTED_VERSION
    if [ $? -eq 0 ]; then
      echo "ERROR! Fabric Docker image version of $DOCKER_IMAGE_VERSION does not match this newer version of BYFN and is unsupported. Either move to a later version of Fabric or checkout an earlier version of fabric-samples."
      exit 1
    fi
  done
}

# a="test"
# b=$(docker images)
# echo ${b}

# displayLogo
# checkImages
# clearContainers 
clearAllContainers
removeUnwantedImages
rm -rf ../volumes
# checkPrereqs