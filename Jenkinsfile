pipeline {
    agent {
        label 'docker'  // Runs on agent with Docker
    }
    
    environment {
        // Use credentials stored in Jenkins
        GITHUB_TOKEN = credentials('github-pat')
    }
    
    stages {
        stage('Checkout') {
            steps {
                checkout scm
                sh 'git log -1'
            }
        }
        
        stage('Build Info') {
            steps {
                script {
                    echo "Branch: ${env.GIT_BRANCH}"
                    echo "Commit: ${env.GIT_COMMIT}"
                    echo "Build Number: ${env.BUILD_NUMBER}"
                }
            }
        }
        
        stage('Test') {
            steps {
                sh '''
                    echo "Running tests..."
                    # Add your test commands here
                '''
            }
        }
    }
    
    post {
        success {
            echo "✅ Build successful!"
        }
        failure {
            echo "❌ Build failed!"
        }
    }
}
