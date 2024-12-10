cd $(dirname "$0")

mkdir -p log
nohup ./jamger config.yml  >> ./log/jamger.log 2>&1 &