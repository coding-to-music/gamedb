if ($('#price-changes-page').length > 0) {

    $('table.table-datatable2').DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[5, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]).attr('data-link', data[7]);

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
                    return '<img src="' + row[8] + '" class="rounded square"><span>' + row[1] + '</span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).addClass('img').attr('data-app-id', 0)
                },
                "orderable": false,
                "searchable": false
            },
            // Release Date
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return row[2];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                }
            },
            // Price
            {
                "targets": 2,
                "render": function (data, type, row) {
                    return '$' + row[3];
                }
            },
            // Discount %
            {
                "targets": 3,
                "render": function (data, type, row) {
                    return row[4] + '%';
                }
            },
            // Price Change
            {
                "targets": 4,
                "render": function (data, type, row) {
                    return '$' + row[5];
                }
            },
            // Time
            {
                "targets": 5,
                "render": function (data, type, row) {
                    return row[6];
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
