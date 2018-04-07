
// https://datatables.net/reference/option/
$('table.table-datatable').DataTable({
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

$('.dataTables_wrapper').removeClass('container-fluid');

var table = $('#DataTables_Table_0');
if (table.length) {
    table = table.DataTable();

    $('input#search').keyup(function () {
        table.search($(this).val()).draw();
    });
}
