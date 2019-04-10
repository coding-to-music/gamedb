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

                                if (row[3][k] === '') {
                                    row[3][k] = 'Unknown App';
                                }

                                apps.push('<a href="/apps/' + k + '">' + row[3][k] + '</a>');
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

                                if (row[4][k] === '') {
                                    row[4][k] = 'Unknown Package';
                                }

                                packages.push('<a href="/packages/' + k + '">' + row[4][k] + '</a>');
                            }
                        }
                    }

                    return packages.join('<br/>');
                },
                "orderable": false
            }
        ]
    });

    function sortByProductName(a, b) {
        if (a.name < b.name)
            return -1;
        if (a.name > b.name)
            return 1;
        return 0;
    }

    const $table = $('table.table-datatable2');
    const dt = $table.DataTable(options);

    websocketListener('changes', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = $.parseJSON(e.data);

            // Loop changes in websocket data and add each one
            if (isIterable(data.Data)) {
                for (const v of data.Data) {
                    addDataTablesRow(options, v, info.length, $table);
                }
            }
        }
    });
}
