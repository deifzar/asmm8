@Library('devops-pipeline-libraries') _

golangPipeline {
  environment             = 'staging'
  repoName                = 'asmm8'
  scmProvider             = 'github' // 'github' or 'gitlab'

  runTest                 = true
  runSAST                 = true // Sonarqube SAST
  runSCA                  = true
  runSBOM                 = true
  runPublish              = true
  
  buildingImage           = 'golang:1.24'

  trivyThreshold          = 'CRITICAL,HIGH'
  trivySkipDirs           = []
  trivySkipFiles          = ['usr/local/bin/dnsx',
                            'usr/local/bin/alterx',
                            'usr/local/bin/subfinder']
  snykThreshold           = 'high'
  snykSkipDirs            = []
  snykSkipFiles           = []
  snykCredentialsId       = 'snyk-pat-jenkins'

  createPullOrMergeRequest = true
  composeStackRepo        = 'gitlab.com/cptm8microservices/cptm8-compose-stack.git' // omit `https://`
  gitCredentialsId        = 'gitlab-pat-jenkins' // gitlab-pat-jenkins || github-app-jenkins
  
  
  sonarqubeCredentialsId  = 'sonarqube-token'  // Jenkins credentials ID for SonarQube token 
  sonarqubeUrl            = 'https://sonarqube-staging.cptm8.net'
  artifactoryCredentialsId  = 'artifactory-pat'   // Jenkins credentials ID for Artifactory token 
  artifactoryUrl            = 'https://trial0ve3le.jfrog.io'
  artifactoryGenericRepo    = 'cptm8-generic'  // e.g., 'cptm8-generic'
  artifactoryDockerRepo     = 'cptm8-docker'  /
}

/* service pipeline mapp
golangPipeline {
    environment               : 'dev',
    repoName                  : null, // 'asmm8' || 'katanam8' ...
    scmProvider               : null, // 'github' or 'gitlab'
    // Main steps:
    runTest                   : false,
    runSAST                   : false, // Sonarqube
    runSCA                    : false,
    runSBOM                   : true,
    runPublish                : false, // Artifactory

    buildingImage                : 'golang:1.23',

    // runTrivySourceScan        : false,
    // runTrivyImageScan         : true,   // Trivy image scan (enabled by default)
    // runTrivyIaCScan           : false,

    // trivy settings
    trivyThreshold            : 'HIGH,CRITICAL', // strings with comma: CRITICAL,HIGH,MEDIUM,LOW
    trivySkipDirs             : [],     // List of directories to skip in Trivy SCA scan
    trivySkipFiles            : [],     // List of files to skip in Trivy SCA scan
    // snyk settings
    snykThreshold             : 'high'  // single string: critical, high, medium, low
    snykSkipDirsOrFiles       : [],     // List of directories or files to skip in Snyk SCA scan

    // Binary publishing config
    createPullOrMergeRequest  : true,
    composeStackRepo          : null, // 'gitlab.com/cptm8microservices/cptm8-compose-stack.git' || github.com/deifzar/cptm8-compose-stack.git
    // Git credentials
    gitCredentialsId          : null,  // Jenkins credentials ID // gitlab-pat-jenkins || github-app-jenkins
    // Snyk credentials
    snykCredentialsId         : null, // snyk-pat-jenkins
    // SonarQube config
    sonarqubeCredentialsId    : null,  // Jenkins credentials ID for SonarQube token
    sonarqubeUrl              : null,  // SonarQube server URL (e.g., 'https://sonar.example.com')
    // Artifactory config
    artifactoryCredentialsId  : null,  // Jenkins credentials ID for Artifactory token 
    artifactoryUrl            : null,
    artifactoryGenericRepo    : null,  // e.g., 'cptm8-generic'
    artifactoryDockerRepo     : null,  // e.g., 'cptm8-docker'
  }
*/
