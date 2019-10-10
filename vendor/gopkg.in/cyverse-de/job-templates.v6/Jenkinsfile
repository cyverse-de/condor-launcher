#!groovy
node('docker') {
    slackJobDescription = "job '${env.JOB_NAME} [${env.BUILD_NUMBER}]' (${env.BUILD_URL})"
    try {
        stage "Build"
        checkout scm

        service = readProperties file: 'service.properties'

        descriptive_version = sh(returnStdout: true, script: 'git describe --long --tags --dirty --always').trim()
        echo descriptive_version

        dockerRepo = "test-${env.BUILD_TAG}"

        sh "docker build --rm -t ${dockerRepo} ."

        dockerTestRunner = "test-${env.BUILD_TAG}"
        dockerTestCleanup = "test-cleanup-${env.BUILD_TAG}"
        try {
            stage "Test"
            try {
                sh "docker run --name ${dockerTestRunner} --rm ${dockerRepo} > test-results.xml"
            } finally {
                junit 'test-results.xml'

                sh "docker run --rm --name ${dockerTestCleanup} -v \$(pwd):/build -w /build alpine rm -r test-results.xml"
            }
        } finally {
            sh returnStatus: true, script: "docker kill ${dockerTestRunner}"
            sh returnStatus: true, script: "docker rm ${dockerTestRunner}"

            sh returnStatus: true, script: "docker kill ${dockerTestCleanup}"
            sh returnStatus: true, script: "docker rm ${dockerTestCleanup}"

            sh returnStatus: true, script: "docker rmi ${dockerRepo}"

            step([$class: 'hudson.plugins.jira.JiraIssueUpdater',
                    issueSelector: [$class: 'hudson.plugins.jira.selector.DefaultIssueSelector'],
                    scm: scm,
                    labels: [ "${service.repo}-${descriptive_version}" ]])
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
