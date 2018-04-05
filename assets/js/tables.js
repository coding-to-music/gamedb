// var $table = $('.table-fix-head');
// $table.floatThead({
//     responsiveContainer: function($table){
//         return $table.closest('.table-responsive');
//     }
// });

// $(document).ready(function () {
//         $('table.table-sorter').tablesorter();
//     }
// );

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
