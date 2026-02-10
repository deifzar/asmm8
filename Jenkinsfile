@Library('microservices-lib') _

servicePipeline {
  serviceName     = 'asmm8'
  buildImage      = 'golang:1.24'
  runCodeScan     = false
  runImageScan    = true
  trivySeverity   = 'CRITICAL,HIGH'
  trivySkipDirs   = []
  trivySkipFiles  = ['usr/local/bin/dnsx',
                    'usr/local/bin/alterx',
                    'usr/local/bin/subfinder']
  deploy          = false
  environments = ['dev']
}

/* service pipeline map
servicePipeline = {
    serviceName       : null,
    dockerfile        : 'Dockerfile',
    imageRegistry     : 'ghcr.io/deifzar',
    runTests          : false,
    runCodeScan       : false,
    runImageScan      : true,   // Trivy image scan (enabled by default)
    trivySkipDirs     : []      // List of directories to skip in Trivy scan
    trivySeverity     : 'HIGH,CRITICAL',
    trivySkipFiles    : [],     // List of files to skip in Trivy scan
    deploy            : false,
    environments      : ['dev'],
    buildImage        : 'golang:1.23', // default Docker image. Other example: node:20
    goBinary          : null   // defaults to serviceName if not set
}
*/