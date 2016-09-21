#!groovy
def repo = "condor-launcher"
def dockerUser = "discoenv"

node('docker') {
    slackJobDescription = "job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})"
    try {
        stage "Build"
        checkout scm

        git_commit = sh(returnStdout: true, script: "git rev-parse HEAD").trim()
        echo git_commit

        dockerRepo = "test-${env.BUILD_TAG}"

        sh "docker build --rm --build-arg git_commit=${git_commit} -t ${dockerRepo} ."


        try {
            stage "Test"
            dockerTestRunner = "test-${env.BUILD_TAG}"
            sh "docker run --rm --name ${dockerTestRunner} --entrypoint 'go' ${dockerRepo} test github.com/cyverse-de/${repo}"

            stage "Docker Push"
            dockerPushRepo = "${dockerUser}/${repo}:${env.BRANCH_NAME}"
            sh "docker tag ${dockerRepo} ${dockerPushRepo}"
            sh "docker push ${dockerPushRepo}"
        } finally {
            sh returnStatus: true, script: "docker kill ${dockerTestRunner}"
            sh returnStatus: true, script: "docker rm ${dockerTestRunner}"

            sh returnStatus: true, script: "docker rmi ${dockerRepo}"
        }
    } catch (InterruptedException e) {
        currentBuild.result = "ABORTED"
        slackSend color: 'warning', message: "ABORTED: ${slackJobDescription}"
        throw e
    } catch (e) {
        currentBuild.result = "FAILED"
        sh "echo ${e}"
        slackSend color: 'danger', message: "FAILED: ${slackJobDescription}"
        throw e
    }
}
