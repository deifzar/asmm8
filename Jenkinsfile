@Library('microservices-lib') _

servicePipeline {
  serviceName = 'asmm8'
  buildImage   = 'golang:1.24'
  runCodeScan     = false
  deploy      = false
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
    trivySeverity     : 'HIGH,CRITICAL',
    deploy            : false,
    environments      : ['dev'],
    buildImage        : 'golang:1.23', // default Docker image. Other example: node:20
    goBinary          : null   // defaults to serviceName if not set
}
*/