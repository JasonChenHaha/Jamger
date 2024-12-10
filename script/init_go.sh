root=`pwd`
exclude_paths=("./project" "./test", "./script")
temp_file=$(mktemp)

find . $(for path in "${exclude_paths[@]}"; do echo -n "-path $path -prune -o "; done) -path '*/.*' -prune -o ! -path '.' -type d -print | while read dir; do
    echo "../../$dir" >> $temp_file
    cd $root/${dir#./}
    if [[ ! -f ./go.mod ]]; then
        pwd
        go mod init j$(basename $dir)
    fi
    go mod tidy
    cd $root
done

while IFS= read -r dir; do
    all_dirs="$all_dirs $dir"
done < "$temp_file"

find ./project -maxdepth 1 ! -path './project' -type d -print | while read dir; do
    project=$(basename $dir)
    cd $root/${dir#./}
    if [[ ! -f ./go.mod ]]; then
        go mod init $project
        go work init $all_dirs
        go work use "./"
    fi
    go mod tidy
    find . -path '*/.*' -prune -o ! -path '.' -type d -print | while read dir2; do
        cd $root/project/$project/${dir2#./}
        if [[ ! -f ./go.mod ]]; then
            go mod init $project$(basename $dir2)
            go mod tidy
            cd $root/${dir#./}
            go work use $dir2
        else
            go mod tidy
        fi
    done
    cd $root
done