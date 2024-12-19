temp_file=$(mktemp)

find ./project -maxdepth 1 ! -path './project' -type d -print | while read dir; do
    project=$(basename $dir)
    echo "SVR_`echo $project | tr 'a-z' 'A-Z'` = \"$project\"" >> $temp_file
done

str='SVR_BEGIN = "nil"\n'
while IFS= read -r dir; do
    str="$str    $dir\n"
done < "$temp_file"
str="$str    SVR_END = \"nil\""

perl -0777 -pi -e "s/SVR_BEGIN.*SVR_END *= \"nil\"/$str/s" ./global/global.go
