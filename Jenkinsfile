/*
 *  Copyright (C) 2021-2023 Red Hat, Inc.
 *
 *  Licensed under the Apache License, Version 2.0 (the "License");
 *  you may not use this file except in compliance with the License.
 *  You may obtain a copy of the License at
 *
 *          http://www.apache.org/licenses/LICENSE-2.0
 *
 *  Unless required by applicable law or agreed to in writing, software
 *  distributed under the License is distributed on an "AS IS" BASIS,
 *  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *  See the License for the specific language governing permissions and
 *  limitations under the License.
 */

// This Jenkinsfile is used as CI script for indy-tests project

def build_image="quay.io/factory2/spmm-jenkins-agent-go:latest"
pipeline {
  agent {
    kubernetes {
      cloud params.JENKINS_AGENT_CLOUD_NAME
      // label "jenkins-slave-${UUID.randomUUID().toString()}"
      serviceAccount "jenkins"
      defaultContainer 'jnlp'
      yaml """
      apiVersion: v1
      kind: Pod
      metadata:
        labels:
          app: "jenkins-${env.JOB_BASE_NAME}"
          indy-pipeline-build-number: "${env.BUILD_NUMBER}"
      spec:
        containers:
        - name: jnlp
          image: ${build_image}
          imagePullPolicy: Always
          tty: true
          env:
          - name: HOME
            value: /home/jenkins
          - name: GOROOT
            value: /usr/lib/golang
          - name: GOPATH
            value: /home/jenkins/gopath
          - name: GOPROXY
            value: https://proxy.golang.org
          resources:
            requests:
              memory: 4Gi
              cpu: 2000m
            limits:
              memory: 8Gi
              cpu: 4000m
          workingDir: /home/jenkins
      """
    }
  }
  options {
    //timestamps()
    timeout(time: 360, unit: 'MINUTES')
  }
  environment {
    PIPELINE_NAMESPACE = readFile('/run/secrets/kubernetes.io/serviceaccount/namespace').trim()
    PIPELINE_USERNAME = sh(returnStdout: true, script: 'id -un').trim()
  }
  stages {
    stage('Show golang environment'){
      steps{
        script{
          sh 'go version'
          sh 'go env'
        }
      }
    }
    stage('Build & Test') {
      steps {
        sh 'make test'
      }
    }
  }
  post {
    success {
      script {
        echo "SUCCEED"
      }
    }
    failure {
      script {
        echo "FAILED"
      }
    }
  }
}


