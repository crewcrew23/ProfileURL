pipeline {
    agent any
    
    parameters {
        string(name: 'ENV', defaultValue: 'local', description: 'Environment')
        string(name: 'ADDR', defaultValue: ':5500', description: 'Address')
        string(name: 'STORAGE_PATH', defaultValue: './storage/url_profile.db', description: 'Path to DB')
        string(name: 'TOKEN_TTL', defaultValue: '1h', description: 'Token lifetime')
        password(name: 'SECRET', defaultValue: '', description: 'JWT secret key')
    }

    environment {
        ENV = "${params.ENV}"
        ADDR = "${params.ADDR}"
        STORAGE_PATH = "${params.STORAGE_PATH}"
        TOKEN_TTL = "${params.TOKEN_TTL}"
        SECRET = "${params.SECRET}"
    }

    stages {
        stage('Prepare config') {
            steps {
                echo 'Generating config.yml from config.tpl.yml...'
                sh '''
                    ls  -l
                    ls -b ./config
                    envsubst < ./config/config.tpl.yml > ./config/config.yml
                    echo "Generated config.yml:"
                    cat ./config/config.yml
                '''
            }
        }

        stage('Prepare dirs') {
            steps {
                sh '''
                    mkdir -p bin
                    mkdir -p storage
                '''
            }
        }

        stage('Build migrator') {
            steps {
                sh 'go build -o ./bin/migrator ./cmd/migrator/main.go'
            }
        }

        stage('Run migrator') {
            steps {
                sh './bin/migrator --storage=$STORAGE_PATH --migration-path=./migrations'
            }
        }

        stage('Build server') {
            steps {
                sh 'go build -o ./bin/server ./cmd/app/main.go'
            }
        }
    }

    post {
        success {
            echo 'Build completed successfully.'
        }
        failure {
            echo 'Build failed.'
        }
    }
}