if ($('#price-changes-page').length > 0) {

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[5, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]);
            $(row).attr('data-link', data[5]);

            if (data[7] > 0) {
                $(row).addClass('table-danger');
            } else if (data[7] < 0) {
                $(row).addClass('table-success');
            }
        },
        "columnDefs": [
            // App/Package Name
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return '<img src="' + row[4] + '" class="rounded square" alt="' + row[3] + '"><span>' + row[3] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img').attr('data-app-id', 0)
                },
                "orderable": false
            },
            // Before
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[6];
                },
                "orderable": false
            },
            // After
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return row[7];
                },
                "orderable": false
            },
            // Discount
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[8] + ' - ' + row[9] + '%';
                },
                "orderable": false
            },
            // Time
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[10] + '" data-livestamp="' + row[11] + '">' + row[10] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                }
            }
        ]
    }));

    // var $hideRed = $('#hide-red');
    // var $hideGreen = $('#hide-green');
    // var $hideApps = $('#hide-apps');
    // var $hidePackages = $('#hide-packages');
    // var $hideOwned = $('#hide-owned');
    //
    // $.fn.dataTable.ext.search.push(
    //     function (settings, searchData, index, rowData, counter) {
    //
    //         var change = Number(searchData[5].replace(/[^0-9\.-]+/g, ""));
    //
    //         if ($hideRed.is(':checked') && change > 0) {
    //             return false;
    //         }
    //
    //         if ($hideGreen.is(':checked') && change < 0) {
    //             return false;
    //         }
    //
    //         if ($hideApps.is(':checked')) {
    //
    //             var appID = table
    //                 .row(index)         //get the row to evaluate
    //                 .nodes()                //extract the HTML - node() does not support to$
    //                 .to$()                  //get as jQuery object
    //                 // .find('td[data-label]') //find column with data-label
    //                 // .data('label');         //get the value of data-label
    //                 .attr('data-app-id');
    //
    //             if (appID > 0) {
    //                 return false;
    //             }
    //         }
    //
    //         if ($hidePackages.is(':checked')) {
    //
    //             var packageID = table
    //                 .row(index)         //get the row to evaluate
    //                 .nodes()                //extract the HTML - node() does not support to$
    //                 .to$()                  //get as jQuery object
    //                 // .find('td[data-label]') //find column with data-label
    //                 // .data('label');         //get the value of data-label
    //                 .attr('data-package-id');
    //
    //             if (packageID > 0) {
    //                 return false;
    //             }
    //         }
    //
    //         return true;
    //     }
    // );
    //
    // $('#hide-red, #hide-green, #hide-apps, #hide-packages, #hide-owned').change(function () {
    //
    //     $('#DataTables_Table_0').DataTable().draw();
    //
    // })


}
