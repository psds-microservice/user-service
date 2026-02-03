$b = [System.IO.File]::ReadAllBytes("$PSScriptRoot\..\fd.pb")
# Read varint at position $i, advance $i past it, return value
function Read-Varint {
    param([array]$bytes, [ref]$idx)
    $v = 0
    $shift = 0
    while ($idx.Value -lt $bytes.Length) {
        $byte = $bytes[$idx.Value]
        $idx.Value++
        $v += ($byte -band 0x7f) -shl $shift
        if ($byte -lt 128) { return $v }
        $shift += 7
    }
    return $v
}
$i = 0
# Skip to first 0x0a (field 1, start of first file)
while ($i -lt $b.Length -and $b[$i] -ne 0x0a) { $i++ }
$i++  # past 0x0a
$len1 = Read-Varint $b ([ref]$i)
$i += $len1  # skip first file content
# Find second 0x0a (start of second file)
while ($i -lt $b.Length -and $b[$i] -ne 0x0a) { $i++ }
$i++  # past 0x0a
$len2 = Read-Varint $b ([ref]$i)
$end = [Math]::Min($i + $len2, $b.Length)
$chunk = $b[$i..($end-1)]
$s = ""
foreach ($x in $chunk) { $s += "\x" + ("{0:x2}" -f $x) }
$goLiteral = "`"" + $s + "`""
Set-Content -Path "$PSScriptRoot\..\pkg\gen\user_service\rawdesc.txt" -Value $goLiteral -NoNewline
