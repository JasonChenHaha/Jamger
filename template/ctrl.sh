cd $(dirname "$0")

if [ "$1" == "start" ]; then
    while IFS= read -r line || [ -n "$line" ]; do
        read -a arr <<< $line
        if [[ ! "$line" =~ ^[[:space:]]*$ ]]; then
            zone="${arr[0]#*=}"
            group="${arr[1]#*=}"
            index="${arr[2]#*=}"
            tcp="${arr[3]#*=}"
            kcp="${arr[4]#*=}"
            web="${arr[5]#*=}"
            http="${arr[6]#*=}"
            grpc="${arr[7]#*=}"
            server=$group-$index
            if [[ `ps ux | grep $server | grep -v grep | grep -v 'tail' | wc -l` -eq 0 ]]; then
                cd ./$group
                sed -e "s/\$ZONE/\"$zone\"/g" -e "s/\$GROUP/\"$group\"/g" -e "s/\$INDEX/\"$index\"/g" -e "s/\$TCP/$tcp/g" -e "s/\$KCP/$kcp/g" -e "s/\$WEB/$web/g" -e "s/\$HTTP/$http/g" -e "s/\$GRPC/$grpc/g" ../config.yml > .config.yml
                exec -a "$server" ./$group ./.config.yml >> ./log/$server.log 2>&1 &
                cd ..
            fi
        fi
    done < "./serverList"
elif [ "$1" == "stop" ]; then
    while IFS= read -r line || [ -n "$line" ]; do
        read -a arr <<< $line
        if [[ ! "$line" =~ ^[[:space:]]*$ ]]; then 
            group="${arr[1]#*=}"
            index="${arr[2]#*=}"
            server=$group-$index
            if [[ `ps ux | grep $server | grep -v grep | grep -v 'tail' | wc -l` -gt 0 ]]; then
                ps ux | grep $server | grep -v grep | grep -v 'tail' | awk '{print $2}' | xargs kill -2
                rm -f ./$group/.config.yml
            fi
        fi
    done < "./serverList"
elif [ "$1" == "info" ]; then
    while IFS= read -r line || [ -n "$line" ]; do
        read -a arr <<< $line
        if [[ ! "$line" =~ ^[[:space:]]*$ ]]; then
            group="${arr[1]#*=}"
            index="${arr[2]#*=}"
            server=$group-$index
            ps ux | grep $server | grep -v grep | grep -v 'tail' || true
        fi
    done < "./serverList"
fi