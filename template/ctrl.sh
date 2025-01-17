cd $(dirname "$0")

if [ "$1" == "start" ]; then
    for file in *.yml; do
        name=$(grep -oP '^name:\s*\K.*' "$file")
        index=$(grep -oP '^index:\s*\K.*' "$file")
        server=$name-$index
        if [[ `ps ux | grep $server | grep -v grep | grep -v 'tail' | wc -l` -eq 0 ]]; then
            exec -a "$server" ./$name ./$file >> ./log/$server.log 2>&1 &
        fi
    done
elif [ "$1" == "stop" ]; then
    for file in *.yml; do
        name=$(grep -oP '^name:\s*\K.*' "$file")
        index=$(grep -oP '^index:\s*\K.*' "$file")
        server=$name-$index
        if [[ `ps ux | grep $server | grep -v grep | grep -v 'tail' | wc -l` -gt 0 ]]; then
            ps ux | grep $server | grep -v grep | grep -v 'tail' | awk '{print $2}' | xargs kill -2
        fi
    done
elif [ "$1" == "info" ]; then
    for file in *.yml; do
        name=$(grep -oP '^name:\s*\K.*' "$file")
        index=$(grep -oP '^index:\s*\K.*' "$file")
        server=$name-$index
        ps ux | grep $server | grep -v grep | grep -v 'tail' || true
    done
fi