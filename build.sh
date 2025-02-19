#!/bin/bash

RED="\e[31m"
GREEN="\e[32m"
BLUE="\e[34m"
YELLOW="\e[33m"
NC="\e[0m"

REDBGR='\033[0;41m'
NCBGR='\033[0m'

########## CONFIG ##########
DOCKER_REGISTRY="bonavadeur"
IMAGE="katyusha" # docker.io/{DOCKER_REGISTRY}/{IMAGE}
NAMESPACE="knative-serving"
component="activator"
OPTION=$1
############################

logSuccess() { echo -e "$GREEN-----$1-----$NC";}
logInfo() { echo -e "$BLUE-----$1-----$NC";}
logError() { echo -e "$RED-----$1-----$NC";}
logStage() { echo -e "$YELLOW###############---$1---###############$NC";}

dockerBuild() {
    compress=$1
    mode=$2
    
    logStage "Go build binary"
    CGO_ENABLED=0 go build -ldflags="-s -w" -o ./$IMAGE ./cmd/$IMAGE

    logStage "Compress binary"
    if [ $compress == "fast" ]; then
        upx -1 $IMAGE
    elif [ $compress == "high" ]; then
        upx -9 $IMAGE
    fi

    logStage "Docker build image"
    docker rmi $DOCKER_REGISTRY/$IMAGE:dev
    docker build --no-cache -t $DOCKER_REGISTRY/$IMAGE:dev .
    docker save -o $IMAGE.tar $DOCKER_REGISTRY/$IMAGE:dev

    if [ "$mode" == "mnode" ]; then
        scp $IMAGE.tar cloud-node:~/
        scp $IMAGE.tar edge-node:~/
    fi

    sudo crictl rmi docker.io/$DOCKER_REGISTRY/$IMAGE:dev
    sudo ctr -n=k8s.io images import $IMAGE.tar
    if [ "$mode" == "mnode" ]; then
        ssh cloud-node "crictl rmi docker.io/$DOCKER_REGISTRY/$IMAGE:dev"
        ssh cloud-node "ctr -n=k8s.io images import $IMAGE.tar"
        ssh edge-node "crictl rmi docker.io/$DOCKER_REGISTRY/$IMAGE:dev"
        ssh edge-node "ctr -n=k8s.io images import $IMAGE.tar"
    fi

    logStage "Clean up"
    rm -rf ./$IMAGE
    rm -rf $IMAGE.tar
    if [ "$mode" == "mnode" ]; then
        ssh cloud-node "rm -rf $IMAGE.tar"
        ssh edge-node "rm -rf $IMAGE.tar"
    fi
}

pushDockerImage() {
    logStage "pushing image to Docker Hub"
    tag=$1
    CONTAINER_REGISTRY="docker.io"/$DOCKER_REGISTRY
    docker tag $CONTAINER_REGISTRY/$IMAGE:dev $CONTAINER_REGISTRY/$IMAGE:$tag
    docker push $CONTAINER_REGISTRY/$IMAGE:$tag
    docker rmi $CONTAINER_REGISTRY/$IMAGE:$tag
}

deployNewVersion() {
    logStage "remove current Pod"
    pods=($(kubectl -n $NAMESPACE get pod | grep $component | awk '{print $1}'))
    for pod in ${pods[@]}
    do
        kubectl -n $NAMESPACE delete pod/$pod &
    done
}

logPod() {
    sleep 1
    pods=($(kubectl -n $NAMESPACE get pod | grep $component | grep Running | awk '{print $1}'))
    while [ "${pods[0]}" == "" ];
    do
        sleep 1
        pods=($(kubectl -n $NAMESPACE get pod | grep $component | grep Running | awk '{print $1}'))
    done
    echo "pod:"${pods[0]}
    kubectl -n $NAMESPACE wait --for=condition=ready pod ${pods[0]} > /dev/null 2>&1
    clear
    endTime=`date +%s`
    logInfo "KoBuild time was $koBuildTime seconds."
    logInfo "Build time was `expr $endTime - $startTime` seconds."
    logStage "$IMAGE logs"
    echo "pod:"${pods[0]}
    kubectl -n $NAMESPACE logs ${pods[0]} -f
}
#
#
#
#
#
#
#
#
#
#
clear
echo -e "$REDBGR このスクリプトはボナちゃんによって書かれています $NCBGR"

startTime=`date +%s`

if [ $OPTION == "ful" ]; then
    # ./build.sh ful snode
    # ./build.sh ful mnode
    dockerBuild fast $2
    deployNewVersion
    logPod
elif [ $OPTION == "push" ]; then
    # ./build.sh push snode
    # ./build.sh push mnode
    dockerBuild "high" $2
    pushDockerImage $2
    deployNewVersion
    sleep 1
elif [ $OPTION == "log" ]; then
    deployNewVersion
    logPod
fi
