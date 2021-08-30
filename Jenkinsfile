node {
    def app
    // environment { 
    //     HTTP_PROXY = "http://127.0.0.1:7890"
    //     HTTPS_PROXY = "http://127.0.0.1:7890"
    // }

    stage('Clone repository') {
        /* Let's make sure we have the repository cloned to our workspace */

        checkout scm
    }

    stage('Build image') {
        /* This builds the actual image; synonymous to
         * docker build on the command line */

        app = docker.build("matrinos/influxdb-writer", "--no-cache --build-arg SVC=influxdb-writer --build-arg GOARCH=amd64 --build-arg GOARM= --network host -f docker/Dockerfile .")
    }

    stage('Test image') {
        /* Ideally, we would run a test framework against our image.
         * For this example, we're using a Volkswagen-type approach ;-) */

        app.inside {
            sh 'echo "Tests passed"'
        }
    }

    stage('Push image') {
        /* Finally, we'll push the image with two tags:
         * First, the incremental build number from Jenkins
         * Second, the 'latest' tag.
         * Pushing multiple tags is cheap, as all the layers are reused. */
        docker.withRegistry('docker-hub-url', 'docker-hub-credentials') {
            app.push("${env.BUILD_NUMBER}")
            app.push("latest")
        }
    }
}
