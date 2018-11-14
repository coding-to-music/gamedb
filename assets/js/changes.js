if ($('#changes-page').length > 0) {

    const columnDefs = [
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
                if (isIterable(Array)) {
                    for (const v of row[3]) {
                        if (v.name === '') {
                            apps.push('Unknown App');
                        } else {
                            apps.push('<a href="/games/' + v.id + '">' + v.name + '</a>');
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
                if (isIterable(row[4])) {
                    for (const v of row[4]) {
                        if (v.name === '') {
                            packages.push('Unknown Package');
                        } else {
                            packages.push('<a href="/packages/' + v.id + '">' + v.name + '</a>');
                        }
                    }
                }


                return packages.join('<br/>');
            },
            "orderable": false
        }
    ];

    const $table = $('table.table-datatable2');

    const dt = $table.DataTable($.extend(true, {}, dtDefaultOptions, {
        "order": [[0, 'desc']],
        "createdRow": function (row, data, dataIndex) {
            $(row).attr('data-id', data[0]).attr('data-link', '/changes/' + data[0]);
        },
        "columnDefs": columnDefs
    }));

    websocketListener('changes', function (e) {

        const info = dt.page.info();
        if (info.page === 0) { // Page 1

            const data = $.parseJSON(e.data);

            // Loop changes in websocket data and add each one
            if (isIterable(data.Data)) {
                for (const v of data.Data) {
                    addDataTablesRow(columnDefs, v, info.length, $table);
                }
            }
        }
    })
}
