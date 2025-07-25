root=`pwd`
exclude_paths=("./server" "./out" "./script" "./template" "./test")
declare -A keys
tmpFile=$(mktemp)
tmpFile2=$(mktemp)

find . $(for path in "${exclude_paths[@]}"; do echo -n "-path $path -prune -o "; done) -path '*/.*' -prune -o ! -path '.' -type d -print | while read dir; do
    echo "../../$dir" >> $tmpFile
    echo "../$dir" >> $tmpFile2
    cd $root/${dir#./}
    if [[ ! -f ./go.mod ]]; then
        go mod init j$(basename $dir)
    fi
    go mod tidy
    cd $root
done

while IFS= read -r dir; do
    all_dirs="$all_dirs $dir"
done < "$tmpFile"

while IFS= read -r dir; do
    all="$all $dir"
    keys[$(basename $dir)]=1
done < "$tmpFile2"

find ./server -maxdepth 1 ! -path './server' -type d -print | while read dir; do
    server=$(basename $dir)
    cd $root/${dir#./}
    if [[ ! -f ./go.mod ]]; then
        go mod init $server
    fi
    go mod tidy
    rm -f go.work go.work.sum
    go work init $all_dirs
    go work use "./"
    find . -path '*/.*' -prune -o ! -path '.' -type d -print | while read dir2; do
        cd $root/server/$server/${dir2#./}
        if [[ ! -f ./go.mod ]]; then
            key=$(basename $dir2)
            if [[ -n "${keys[$key]}" ]]; then
                go mod init j$(basename $dir2)2
            else
                go mod init j$(basename $dir2)
            fi
            go mod tidy
        fi
        cd $root/${dir#./}
        go work use $dir2
    done
    cd $root
done

cd $root/test
if [[ ! -f ./go.mod ]]; then
    go mod init test
    go work init $all
    go work use "./"
fi
go mod tidy
find . -path '*/.*' -prune -o ! -path '.' -type d -print | while read dir; do
    cd $root/test/${dir#./}
    if [[ ! -f ./go.mod ]]; then
        go mod init $server$(basename $dir)
        go mod tidy
        cd $root/test
        go work use $dir
    else
        go mod tidy
    fi
done
cd $root

rm -f $tmpFile
rm -f $tmpFile2