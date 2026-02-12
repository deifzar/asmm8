@Library('microservices-lib') _

servicePipeline {
  scmProvider             = 'gitlab' // 'github' or 'gitlab'
  serviceName             = 'asmm8'
  buildImage              = 'golang:1.24'
  runTests                = true
  runSASTScan             = true // Sonarqube SAST
  runTrivySourceScan      = true
  runTrivyImageScan       = true
  runTrivyIaCScan         = true
  trivySeverity           = 'CRITICAL,HIGH'
  trivySkipDirs           = []
  trivySkipFiles          = ['usr/local/bin/dnsx',
                            'usr/local/bin/alterx',
                            'usr/local/bin/subfinder']
  deploy                  = false
  environments            = ['dev']
  composeStackRepo        = 'gitlab.com/cptm8microservices/cptm8-compose-stack.git' // omit `https://`
  gitCredentialsId        = 'gitlab-app-jenkins' // gitlab-app-jenkins || github-app-jenkins
  sonarqubeUrl            = 'https://sonarqube-cptm8net.spaincentral.cloudapp.azure.com'
  sonarqubeCredentialsId  = 'sonarqube-token'  // Jenkins credentials ID for SonarQube token
}

/* service pipeline mapp
def servicePipeline = [
    scmProvider             : null, // 'github' or 'gitlab'
    serviceName             : null,
    dockerfile              : 'dockerfile',
    imageRegistry           : null,
    runTests                : false,
    runSASTScan             : false, // Sonarqube
    runTrivySourceScan      : false,
    runTrivyImageScan       : true,   // Trivy image scan (enabled by default)
    runTrivyIaCScan         : false,
    trivySeverity           : 'HIGH,CRITICAL',
    trivySkipDirs           : [],     // List of directories to skip in Trivy scan
    trivySkipFiles          : [],     // List of files to skip in Trivy scan
    deploy                  : false,
    environments            : ['dev'],
    buildImage              : 'golang:1.23',
    goBinary                : null,   // defaults to serviceName if not set
    // Binary publishing config
    publishBinary           : true,
    composeStackRepo        : null, // 'gitlab.com/cptm8microservices/cptm8-compose-stack.git' || github.com/deifzar/cptm8-compose-stack.git
    gitCredentialsId        : null,  // Jenkins credentials ID
    // SonarQube config
    sonarqubeUrl            : null,  // SonarQube server URL (e.g., 'https://sonar.example.com')
    sonarqubeCredentialsId  : null  // Jenkins credentials ID for SonarQube token
  ]
*/
