const $lockIcon = '<i class="fa fa-lock text-muted" data-toggle="tooltip" data-placement="left" title="Private"></i>';

// Local search
const $searchField = $('input#search');
$searchField.on('keyup', function (e) {
    $dataTables.DataTable().search($(this).val()).draw();
});

$searchField.on('keyup', function (e) {
    if ($(this).val() && e.key === "Escape") {
        $(this).val('');
        $dataTables.DataTable().search($(this).val()).draw();
        $dataTables2.DataTable().search($(this).val()).draw();
    }
});

//
function addDataTablesRow(options, data, limit, $table) {

    let $row = $('<tr class="fade-green" />');
    options.createdRow($row[0], data, null);

    if (isIterable(options.columnDefs)) {
        for (const v of options.columnDefs) {

            let value = data[v];

            if ('render' in v) {
                value = v.render(null, null, data);
            }

            const $td = $('<td />').html(value);

            if ('createdCell' in v) {
                v.createdCell($td, null, data, null, null);
            }

            $td.find('[data-livestamp]').html('a few seconds ago');

            $row.append($td);
        }
    }


    $table.prepend($row);

    $table.find('tbody tr').slice(limit).remove();

    observeLazyImages($row.find('img[data-lazy]'));
}
