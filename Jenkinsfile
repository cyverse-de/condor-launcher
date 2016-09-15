#!groovy
def repo = "condor-launcher"
def dockerUser = "discoenv"

node {
    stage "Build"
    checkout scm

    sh 'git rev-parse HEAD > GIT_COMMIT'
    git_commit = readFile('GIT_COMMIT').trim()
    echo git_commit

    dockerRepo = "${dockerUser}/${repo}:${env.BRANCH_NAME}"

    sh "docker build --rm --build-arg git_commit=${git_commit} -t ${dockerRepo} ."


    try {
        stage "Test"
        dockerTestRunner = "test-${env.BUILD_TAG}"
        sh "docker run --rm --name ${dockerTestRunner} --entrypoint 'go' ${dockerRepo} test github.com/cyverse-de/${repo}"
    } finally {
        sh returnStatus: true, script: "docker kill ${dockerTestRunner}"
        sh returnStatus: true, script: "docker rm ${dockerTestRunner}"
    }


    stage "Docker Push"
    sh "docker push ${dockerRepo}"
}
