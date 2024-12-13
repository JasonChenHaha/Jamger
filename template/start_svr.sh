cd $(dirname "$0")

if [[ `ps ux | grep './jamger ' | grep -v grep | wc -l` -eq 0 ]]; then
    mkdir -p log
    nohup ./jamger config.yml  >> ./log/jamger.log 2>&1 &
fi