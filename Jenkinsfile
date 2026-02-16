@Library('devops-pipeline-libraries') _

golangPipeline {
  environment             = 'dev'
  repoName                = 'asmm8'
  scmProvider             = 'gitlab' // 'github' or 'gitlab'

  runTest                 = true
  runSAST                 = true // Sonarqube SAST
  runSCA                  = true
  runSBOM                 = true
  runDeployment           = true
  
  buildingImage           = 'golang:1.24'

  scaSeverity             = 'CRITICAL,HIGH'
  trivySkipDirs           = []
  trivySkipFiles          = ['usr/local/bin/dnsx',
                            'usr/local/bin/alterx',
                            'usr/local/bin/subfinder']
  snykSkipDirs            = []
  snykSkipFiles           = ['usr/local/bin/dnsx',
                            'usr/local/bin/alterx',
                            'usr/local/bin/subfinder']
  createPullOrMergeRequest = true
  composeStackRepo        = 'gitlab.com/cptm8microservices/cptm8-compose-stack.git' // omit `https://`
  gitCredentialsId        = 'gitlab-pat-jenkins' // gitlab-pat-jenkins || github-app-jenkins
  snykCredentialsId       = 'snyk-pat-jenkins'
  
  sonarqubeCredentialsId  = 'sonarqube-token'  // Jenkins credentials ID for SonarQube token 
  sonarqubeUrl            = 'https://sonarqube-cptm8net.spaincentral.cloudapp.azure.com'
  
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
    runDeployment             : false,

    buildingImage                : 'golang:1.23',
    
    // runTrivySourceScan        : false,
    // runTrivyImageScan         : true,   // Trivy image scan (enabled by default)
    // runTrivyIaCScan           : false,

    scaSeverity             : 'HIGH,CRITICAL',
    // trivy settings    
    trivySkipDirs             : [],     // List of directories to skip in Trivy SCA scan
    trivySkipFiles            : [],     // List of files to skip in Trivy SCA scan
    // snyk settings
    snykSkipDirs              : [],     // List of directories to skip in Snyk SCA scan
    snykSkipFiles             : [],     // List of files to skip in Snyk SCA scan
    
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
  }
*/
