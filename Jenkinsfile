def repo = "condor-launcher"
def dockerUser = "discoenv"

node {
    stage "Build"
    git url: "https://github.com/cyverse-de/${repo}", branch: "${env.BRANCH_NAME}"

    sh 'git rev-parse HEAD > GIT_COMMIT'
    git_commit = readFile('GIT_COMMIT').trim()
    echo git_commit

    dockerRepo = "${dockerUser}/${repo}:${env.BRANCH_NAME}"

    sh "docker build --rm --build-arg git_commit=${git_commit} -t ${dockerRepo} ."


    stage "Test"
    sh "docker run --rm --entrypoint 'go' ${dockerRepo} test github.com/cyverse-de/${repo}"


    stage "Docker Push"
    sh "docker push ${dockerRepo}"
}
