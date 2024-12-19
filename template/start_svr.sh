cd $(dirname "$0")

PROJ_NAME=`basename "$PWD"`
INDEX=`grep index config.yml | sed 's/.*"\(.*\)".*/\1/'`

if [[ `ps ux | grep './jamger ' | grep -v grep | wc -l` -eq 0 ]]; then
    mkdir -p log
    exec -a "$PROJ_NAME-$INDEX" ./jamger config.yml  >> ./log/jamger.log 2>&1 &
fi