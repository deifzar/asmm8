@Library('microservices-lib') _

servicePipeline {
  serviceName        = 'asmm8'
  buildImage         = 'golang:1.24'
  runCodeScan        = false
  runTrivySourceScan = true
  runTrivyImageScan  = true
  runTrivyIaCScan    = true
  trivySeverity      = 'CRITICAL,HIGH'
  trivySkipDirs      = []
  trivySkipFiles     = ['usr/local/bin/dnsx',
                    'usr/local/bin/alterx',
                    'usr/local/bin/subfinder']
  deploy            = false
  environments      = ['dev']
  gitCredentialsId  = 'github-app-jenkins'
}

/* service pipeline map
servicePipeline = {
    serviceName         : null,
    dockerfile          : 'Dockerfile',
    imageRegistry       : 'ghcr.io/deifzar',
    runTests            : false,
    runSASTScan         : false,
    runTrivySourceScan  : false,
    runTrivyImageScan   : true,   // Trivy image scan (enabled by default)
    runTrivyIaCScan     : true,
    trivySeverity       : 'HIGH,CRITICAL',
    trivySkipDirs       : [],     // List of directories to skip in Trivy scan
    trivySkipFiles      : [],     // List of files to skip in Trivy scan
    deploy              : false,
    environments        : ['dev'],
    buildImage          : 'golang:1.23',
    goBinary            : null,   // defaults to serviceName if not set
    // Binary publishing config
    publishBinary       : true,
    composeStackRepo    : 'https://github.com/deifzar/cptm8-compose-stack.git',
    gitCredentialsId    : null  // Jenkins credentials ID
}
*/