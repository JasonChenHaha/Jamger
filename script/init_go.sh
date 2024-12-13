root=`pwd`
exclude_paths=("./project" "./out" "./script" "./template" "./test")
temp_file=$(mktemp)
temp_file2=$(mktemp)

find . $(for path in "${exclude_paths[@]}"; do echo -n "-path $path -prune -o "; done) -path '*/.*' -prune -o ! -path '.' -type d -print | while read dir; do
    echo "../../$dir" >> $temp_file
    echo "../$dir" >> $temp_file2
    cd $root/${dir#./}
    if [[ ! -f ./go.mod ]]; then
        go mod init j$(basename $dir)
    fi
    go mod tidy
    cd $root
done

while IFS= read -r dir; do
    all_dirs="$all_dirs $dir"
done < "$temp_file"

while IFS= read -r dir; do
    all_dirs2="$all_dirs2 $dir"
done < "$temp_file2"

find ./project -maxdepth 1 ! -path './project' -type d -print | while read dir; do
    project=$(basename $dir)
    cd $root/${dir#./}
    if [[ ! -f ./go.mod ]]; then
        go mod init $project
    fi
    go mod tidy
    rm -f go.work go.work.sum
    go work init $all_dirs
    go work use "./"
    find . -path '*/.*' -prune -o ! -path '.' -type d -print | while read dir2; do
        cd $root/project/$project/${dir2#./}
        if [[ ! -f ./go.mod ]]; then
            go mod init $project$(basename $dir2)
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
    go work init $all_dirs2
    go work use "./"
fi
go mod tidy
find . -path '*/.*' -prune -o ! -path '.' -type d -print | while read dir; do
    cd $root/test/${dir#./}
    if [[ ! -f ./go.mod ]]; then
        go mod init $project$(basename $dir)
        go mod tidy
        cd $root/test
        go work use $dir
    else
        go mod tidy
    fi
done
cd $root