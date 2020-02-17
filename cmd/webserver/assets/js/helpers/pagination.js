// https://datatables.net/plug-ins/pagination/extjs
$.fn.dataTableExt.oPagination.gamedb = {

    "fnInit": function (oSettings, nPaging, fnCallbackDraw) {
        $(nPaging).prepend($('<ul class="pagination"></ul>'));
        const ul = $("ul", $(nPaging));
        const nFirst = document.createElement('li');
        const nPrevious = document.createElement('li');
        const nNext = document.createElement('li');

        $(nFirst).append($('<a class="page-link">1</a>'));
        $(nPrevious).append($('<a class="page-link">' + (oSettings.oLanguage.oPaginate.sPrevious) + '</a>'));
        $(nNext).append($('<a class="page-link">' + (oSettings.oLanguage.oPaginate.sNext) + '</a>'));

        nFirst.className = 'paginate_button page-item first active';
        nPrevious.className = 'paginate_button page-item previous';
        nNext.className = 'paginate_button page-item next';

        ul.append(nPrevious);
        ul.append(nFirst);
        ul.append(nNext);

        $(nFirst).on('click', function () {
            oSettings.oApi._fnPageChange(oSettings, 'first');
            fnCallbackDraw(oSettings);
        });

        $(nPrevious).on('click', function () {
            if (!(oSettings._iDisplayStart === 0)) {
                oSettings.oApi._fnPageChange(oSettings, 'previous');
                fnCallbackDraw(oSettings);
            }
        });

        $(nNext).on('click', function () {
            if (!(oSettings.fnDisplayEnd() === oSettings.fnRecordsDisplay() || oSettings.aiDisplay.length < oSettings._iDisplayLength)) {
                oSettings.oApi._fnPageChange(oSettings, 'next');
                fnCallbackDraw(oSettings);
            }
        });

        // Reset dynamically generated pages on length/filter change.
        $(oSettings.nTable).DataTable().on('length.dt', function (e, settings, len) {
            $('li.dynamic_page_item', nPaging).remove();
        });

        $(oSettings.nTable).DataTable().on('search.dt', function (e, settings, len) {
            $('li.dynamic_page_item', nPaging).remove();
        });
    },

    /*
     * Function: oPagination.gamedb.fnUpdate
     * Purpose:  Update the list of page buttons shows
     * Inputs:   object:oSettings - dataTables settings object
     *           function:fnCallbackDraw - draw function which must be called on update
     */
    "fnUpdate": function (oSettings, fnCallbackDraw) {
        if (!oSettings.aanFeatures.p) {
            return;
        }

        /* Loop over each instance of the pager */
        const an = oSettings.aanFeatures.p;
        console.log();
        for (let i = 0, iLen = an.length; i < iLen; i++) {
            const buttons = an[i].getElementsByTagName('li');
            $(buttons).removeClass('active');

            if (oSettings._iDisplayStart === 0) {
                buttons[0].className = 'paginate_buttons page-item disabled previous';
                buttons[buttons.length - 1].className = 'paginate_button page-item enabled next';
            } else {
                buttons[0].className = 'paginate_buttons page-item enabled previous';
            }

            const page = Math.round(oSettings._iDisplayStart / oSettings._iDisplayLength) + 1;
            if (page === buttons.length - 1 && oSettings.aiDisplay.length > 0) {
                const $new = $('<li class="dynamic_page_item page-item active"><span class="page-link">' + page + "</span></li>");
                $(buttons[buttons.length - 1]).before($new);
                $new.on('click', function () {
                    $(oSettings.nTable).DataTable().page(page - 1);

                    fnCallbackDraw(oSettings);
                });
            } else {
                $(buttons[page]).addClass('active');
            }

            if (oSettings.fnDisplayEnd() === oSettings.fnRecordsDisplay() || oSettings.aiDisplay.length < oSettings._iDisplayLength) {
                buttons[buttons.length - 1].className = 'paginate_button disabled next';
            }
        }
    }
};
