// https://datatables.net/reference/option/

$("table.table-datatable").each(function (i) {

    var order = [[0, 'asc']];

    var $column = $(this).find('thead tr th[data-sort]');
    if ($column.length > 0) {

        var index = $column.index();
        var sort = $column.attr('data-sort');

        order = [[index, sort]];
    }

    $(this).DataTable({
        "order": order,
        "paging": false,
        "ordering": true,
        "info": false,
        "searching": true,
        "search": {
            "smart": true
        },
        "autoWidth": false,
        "lengthChange": false,
        "stateSave": false,
        "dom": 't'
    });

});


$('.dataTables_wrapper').removeClass('container-fluid');

var table = $('#DataTables_Table_0');
if (table.length) {
    table = table.DataTable();

    $('input#search').keyup(function () {
        table.search($(this).val()).draw();
    });
}
