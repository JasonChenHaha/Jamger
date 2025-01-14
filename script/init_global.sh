tmpFile=$(mktemp)

find ./server -maxdepth 1 ! -path './server' -type d -print | while read dir; do
    server=$(basename $dir)
    echo "SVR_`echo $server | tr 'a-z' 'A-Z'` = \"$server\"" >> $tmpFile
done

str='SVR_BEGIN = "nil"\n'
while IFS= read -r dir; do
    str="$str    $dir\n"
done < "$tmpFile"
str="$str    SVR_END = \"nil\""

perl -0777 -pi -e "s/SVR_BEGIN.*SVR_END *= \"nil\"/$str/s" ./global/global.go

rm -f $tmpFile