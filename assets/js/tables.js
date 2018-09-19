// https://datatables.net/reference/option/

$("table.table-datatable").each(function (i) {

    var order = [[0, 'asc']];

    // Find default column to sort by
    var $column = $(this).find('thead tr th[data-sort]');
    if ($column.length > 0) {

        var index = $column.index();
        var sort = $column.attr('data-sort');

        order = [[index, sort]];
    }

    // Find
    var disabled = [];
    $(this).find('thead tr th[data-disabled]').each(function (i) {
        disabled.push($(this).index());
    });

    // Init
    $(this).DataTable({
        "processing": true,
        "serverSide": true,
        "ajax": {
            "url": "/free-games/ajax",
            "cache": true,
            "data": function (d) {
                delete d.columns;
            }
        },
        // "columns": [
        //     {"data": "id"},
        //     {"data": "name"},
        //     {"data": "description"},
        //     {"data": "createdAt"},
        //     {"data": "updatedAt"}
        // ],
        "order": order,
        "paging": true,
        "ordering": true,
        "info": false,
        "searching": true,
        "search": {
            "smart": true
        },
        "autoWidth": false,
        "lengthChange": false,
        "stateSave": false,
        //"dom": 't',
        "columnDefs": [
            {
                "targets": disabled,
                "orderable": false,
                "searchable": false
            }
        ]
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
