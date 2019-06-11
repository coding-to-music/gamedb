if ($('#changes-page').length > 0) {

    const options = $.extend(true, {}, dtDefaultOptions, {
        "order": [[1, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-link', data[5]);
        },
        "columnDefs": [
            // Change ID
            {
                "targets": 0,
                "render": function (data, type, row) {
                    return 'Change ' + row[0];
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Date
            {
                "targets": 1,
                "render": function (data, type, row) {
                    return '<span data-toggle="tooltip" data-placement="left" title="' + row[2] + '" data-livestamp="' + row[1] + '"></span>';
                },
                "createdCell": function (td, cellData, rowData, row, col) {
                    $(td).attr('nowrap', 'nowrap');
                },
                "orderable": false
            },
            // Apps
            {
                "targets": 2,
                "render": function (data, type, row) {

                    let apps = [];
                    if (row[3] !== null) {
                        for (let k in row[3]) {
                            if (row[3].hasOwnProperty(k)) {

                                if (row[3][k].name === '') {
                                    row[3][k].name = 'Unknown App';
                                }

                                apps.push('<a href="' + row[3][k].path + '">' + row[3][k].name + '</a>');
                            }
                        }
                    }

                    return apps.join('<br/>');
                },
                "orderable": false
            },
            // Packages
            {
                "targets": 3,
                "render": function (data, type, row) {

                    let packages = [];
                    if (row[4] !== null) {
                        for (let k in row[4]) {
                            if (row[4].hasOwnProperty(k)) {

                                if (row[4][k].name === '') {
                                    row[4][k].name = 'Unknown Package';
                                }

                                packages.push('<a href="' + row[4][k].path + '">' + row[4][k].name + '</a>');
                            }
                        }
                    }

                    return packages.join('<br/>');
                },
                "orderable": false
            }
        ]
    });

    const $table = $('table.table-datatable2');
    const dt = $table.DataTable(options);

    websocketListener('changes', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = $.parseJSON(e.data);

            // Loop changes in websocket data and add each one
            if (isIterable(data.Data)) {

                data.Data.sort(function (a, b) {
                    return Math.sign(a[0] - b[0]);
                });

                for (const v of data.Data) {
                    addDataTablesRow(options, v, info.length, $table);
                }
            }
        }
    });
}
