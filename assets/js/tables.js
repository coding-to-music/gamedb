function addDataTablesRow(columnDefs, data, limit, $table) {

    var $row = $('<tr />');

    for (var i in columnDefs) {
        if (columnDefs.hasOwnProperty(i)) {

            var value = data[i];

            if ('render' in columnDefs[i]) {
                value = columnDefs[i].render(null, null, data);
            }

            var $td = $('<td />').html(value);

            if ('createdCell' in columnDefs[i]) {
                columnDefs[i].createdCell($td[0], null, data, null, null); // todo, this [0] may not be needed
            }

            $row.append($td);
        }
    }

    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();
}
