def repo = "configurate"
def dockerUser = "discoenv"

node {
    stage "Build"
    checkout scm

    dockerRepo = "test-${repo}-${env.BRANCH_NAME}"

    sh "docker build --rm -t ${dockerRepo} ."

    stage "Test"
	sh "docker run --rm ${dockerRepo}"

    stage "Clean"
    sh "docker rmi ${dockerRepo}"
}
