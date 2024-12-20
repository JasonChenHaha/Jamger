tmpFile=$(mktemp)

find ./group -maxdepth 1 ! -path './group' -type d -print | while read dir; do
    group=$(basename $dir)
    echo "SVR_`echo $group | tr 'a-z' 'A-Z'` = \"$group\"" >> $tmpFile
done

str='SVR_BEGIN = "nil"\n'
while IFS= read -r dir; do
    str="$str    $dir\n"
done < "$tmpFile"
str="$str    SVR_END = \"nil\""

perl -0777 -pi -e "s/SVR_BEGIN.*SVR_END *= \"nil\"/$str/s" ./global/global.go

rm -f $tmpFile